FROM bufbuild/buf:1.75.0 AS protobuf-tools

FROM golang:1.26.3-alpine AS build
WORKDIR /src

# Include the shared contract module referenced by model/go.mod's local replace.
ENV GOWORK=off CGO_ENABLED=0
COPY --from=protobuf-tools /usr/local/bin/buf /usr/local/bin/buf
COPY go.mod go.sum ./
COPY model/go.mod model/go.sum ./model/
WORKDIR /src/model
RUN go mod download
WORKDIR /src
COPY buf.yaml buf.gen.yaml ./
COPY proto/ ./proto/
RUN buf generate
WORKDIR /src/model
COPY model/ ./
RUN go build -trimpath -ldflags="-s -w" -o /out/model ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/model /model
EXPOSE 50051
ENTRYPOINT ["/model"]
