FROM golang:1.26.3-alpine AS build
WORKDIR /src

# Build the service independently of local workspace replacements.
ENV GOWORK=off CGO_ENABLED=0
COPY model/go.mod model/go.sum ./model/
WORKDIR /src/model
RUN go mod download
COPY model/ ./
RUN go build -trimpath -ldflags="-s -w" -o /out/model ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/model /model
EXPOSE 50051
ENTRYPOINT ["/model"]
