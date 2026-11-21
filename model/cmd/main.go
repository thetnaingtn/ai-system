package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		slog.Error("model service failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	addr := os.Getenv("MODEL_GRPC_ADDR")
	if addr == "" {
		addr = "127.0.0.1:50051"
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	server := grpc.NewServer()
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	// This reports process health only; application readiness comes with the APIs.
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	slog.Info("model service listening", "address", listener.Addr().String())

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	healthServer.Shutdown()
	drained := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(drained)
	}()
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	select {
	case <-drained:
	case <-timer.C:
		server.Stop()
		<-drained
	}
	if err := <-serveErr; err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return err
	}
	slog.Info("model service stopped")
	return nil
}
