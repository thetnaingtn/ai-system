# 30. Go workspace and model service bootstrap

Date: 2026-11-01
Status: Accepted

## Decision

- Use a Go monorepo with the existing root module reserved for shared contracts
  and a separate `github.com/thetnaingtn/ai-system/model` module. Commit `go.work`
  listing both modules; use `go work sync` to synchronize the workspace build
  list back to module manifests. New services add their own module and workspace
  entry. Each module retains its own dependency manifest and checksums.
- Keep protobuf sources in root `proto/`, organized by service and version.
  Application contracts and code generation remain future implementation work.
- Start the model service at `model/cmd/main.go`. The bootstrap exposes only
  standard gRPC process health and bounded graceful shutdown. Application RPCs,
  authenticated tenant resolution, storage clients, providers, and SDKs follow.
- Store paired sequential SQL migrations in `model/migrations/`. Use
  golang-migrate v4.20.1 with the PostgreSQL driver, installed locally via
  `make tools` and run as a separate Compose migration job before model startup.
  The initial migration creates the model schema; domain tables remain open.
  Migration connections set `search_path=public` to keep migration metadata in a
  stable schema even when the model schema matches the local database username.
- Root `compose.yaml` owns service processes and includes `model/compose.yaml`
  for PostgreSQL, Redis, and migrations. Running the latter directly supports
  host-based service development. Compose is for local development; its fixed
  database credential and plaintext loopback endpoint are local-only defaults.
- The Docker build uses the model module independently (`GOWORK=off`). When
  model imports shared generated contracts, add the root module dependency and
  include it in the build context explicitly.

## Consequences

Go workspace dependency selection is shared, but root `go test ./...` does not
traverse nested modules. Root Make targets explicitly check the model module;
extend them when adding packages to the root module or introducing services.
Compose requires version 2.20.3 or later for `include`. PostgreSQL stores durable
local data in a volume; Redis remains disposable. Health reports only process
availability, not database connectivity or inference readiness. Envoy Gateway
and production authentication/deployment remain governed by ADR-0014/0022.
