# AI system

Go monorepo for the AI platform. The first service is the model service; design
decisions are recorded in [docs/adr](docs/adr/README.md).

The proposed schemas, tables, columns, and constraints are documented in the
[AI platform data model](docs/ai-platform-data-model.md).

```text
go.mod                  Shared contract module (github.com/thetnaingtn/ai-system)
go.work                 Workspace and shared dependency selection
proto/                  Shared protobuf sources, grouped by service/version
compose.yaml            Local service orchestration
model/
  go.mod                Model service module
  cmd/main.go           Service entry point
  Dockerfile            Model service image
  compose.yaml          PostgreSQL, Redis, and migration job
  migrations/           Sequential golang-migrate SQL files
```

Requires Go 1.26.3 and Docker Compose 2.20.3 or later.

## Run with Docker

From the repository root:

```sh
docker compose up --build -d
docker compose logs -f model
docker compose down
```

Compose waits for PostgreSQL and Redis health checks, runs migrations, and then
starts the model service. PostgreSQL data persists across restarts. Redis is
temporary and may lose its contents. `docker compose down -v` also deletes the
local PostgreSQL data.

Local endpoints bind to loopback: gRPC `127.0.0.1:50051`, PostgreSQL
`127.0.0.1:5432`, and Redis `127.0.0.1:6379`. Set `MODEL_GRPC_PORT`,
`MODEL_POSTGRES_PORT`, or `MODEL_REDIS_PORT` to override host ports.
The fixed PostgreSQL user/password (`model` / `model-local`) are development
defaults. Production secret management and Envoy Gateway deployment follow later.

The model process currently exposes only standard gRPC health (`service: ""`),
which reports process availability. It has no application APIs, storage clients,
or inference readiness checks yet. Its plaintext local listener must be replaced
by the authenticated transport boundary before application RPCs are introduced.

## Run the model service on the host

```sh
docker compose -f model/compose.yaml up -d --wait model-postgres model-redis
docker compose -f model/compose.yaml run --rm model-migrate
make run
```

`MODEL_GRPC_ADDR` overrides the process listen address (default
`127.0.0.1:50051`). Container orchestration sets it to `0.0.0.0:50051`.
The process handles SIGINT/SIGTERM with a ten-second graceful shutdown limit.
Stop host-development dependencies with
`docker compose -f model/compose.yaml down`.

## Migrations

Install the pinned [golang-migrate CLI](https://github.com/golang-migrate/migrate)
with PostgreSQL support into the ignored repository `bin/` directory:

```sh
make tools
make migrate-up
make migrate-create name=add_model_catalog
make migrate-down
```

`migrate-down` rolls back one migration. `MODEL_DATABASE_URL` overrides the local
connection URL; adjust it when using a custom PostgreSQL port. Keep
`search_path=public` in migration URLs so the version table stays in `public`
after creating the `model` schema. The initial
migration creates an empty `model` schema. Domain tables will be added when their
design is settled. Migrations run explicitly, independently of the service
process; Compose runs the same pinned tool version automatically at startup.

## Workspace and checks

```sh
make sync
make build
make test
make race
make vet
git diff --check
```

`go.work` selects local modules together; `go work sync` writes selected shared
dependency versions back to each module's `go.mod`. Commit module manifests and
checksums as well as workspace changes. After changing dependencies, run
`go mod tidy` in the owning module and `make sync` from the root.

Root Make targets currently check the model module. Root `go test ./...` does not
include nested modules. Add new service modules with `go work use ./<service>`
and extend the Make targets. Protobuf contracts and generation commands will be
introduced together; see [proto/README.md](proto/README.md).
