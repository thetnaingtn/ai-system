# Shared protobuf contracts

All platform `.proto` sources belong here, grouped by service and API version
(for example, `proto/model/v1/`). Generated Go packages will belong to the root
module, `github.com/thetnaingtn/ai-system`.

The application RPCs and generation toolchain are deferred until the contracts
are defined (ADR-0021). The initial model process uses the standard gRPC health
protocol shipped by grpc-go; no application protobuf files are generated yet.
