MIGRATE_VERSION := v4.20.1
MIGRATE := $(CURDIR)/bin/migrate
MODEL_DATABASE_URL ?= postgres://model:model-local@127.0.0.1:5432/model?sslmode=disable&search_path=public
BUF ?= buf
BUF_VERSION := 1.75.0
BUF_BREAKING_AGAINST ?= .git\#branch=main

.PHONY: build test race vet sync run tools migrate-up migrate-down migrate-create proto-tools proto-toolchain proto-format proto-lint proto-generate proto-check proto-check-generated proto-breaking

build:
	go build ./...
	go build -o bin/model ./model/cmd

test:
	go test ./...
	cd model && go test ./...

race:
	go test -race ./...
	cd model && go test -race ./...

vet:
	go vet ./...
	cd model && go vet ./...

sync:
	go work sync

run:
	go run ./model/cmd

tools:
	GOBIN="$(CURDIR)/bin" go install -tags postgres -ldflags="-X main.Version=$(MIGRATE_VERSION)" github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)

# Buf is already installed separately; Go manages the generator versions in go.mod.
proto-toolchain:
	@test "$$($(BUF) --version)" = "$(BUF_VERSION)" || (echo 'Expected Buf $(BUF_VERSION); update the toolchain pin intentionally when upgrading.' >&2; exit 1)

proto-tools: proto-toolchain
	go tool protoc-gen-go --version
	go tool protoc-gen-go-grpc --version

proto-format: proto-toolchain
	$(BUF) format -w

proto-lint: proto-toolchain
	$(BUF) build -o /dev/null
	$(BUF) lint

proto-generate: proto-toolchain
	$(BUF) generate

proto-check-generated: proto-toolchain
	@set -eu; \
	check_dir=$$(mktemp -d); \
	trap 'rm -rf "$$check_dir"' EXIT; \
	$(BUF) generate --output "$$check_dir"; \
	diff -ru gen/go "$$check_dir/gen/go"

proto-check: proto-lint proto-check-generated
	$(BUF) format --diff --exit-code

proto-breaking: proto-toolchain
	$(BUF) breaking --against '$(BUF_BREAKING_AGAINST)'

migrate-up:
	@"$(MIGRATE)" -path model/migrations -database "$(MODEL_DATABASE_URL)" up

migrate-down:
	@"$(MIGRATE)" -path model/migrations -database "$(MODEL_DATABASE_URL)" down 1

migrate-create:
	@test -n "$(name)" || (echo 'usage: make migrate-create name=describe_change' >&2; exit 1)
	"$(MIGRATE)" create -ext sql -dir model/migrations -seq "$(name)"
