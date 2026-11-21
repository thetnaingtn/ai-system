MIGRATE_VERSION := v4.20.1
MIGRATE := $(CURDIR)/bin/migrate
MODEL_DATABASE_URL ?= postgres://model:model-local@127.0.0.1:5432/model?sslmode=disable&search_path=public
BUF ?= buf
BUF_VERSION := 1.75.0
BUF_BREAKING_AGAINST ?= .git\#branch=main

.PHONY: build test race vet sync run tools migrate-up migrate-down migrate-create proto-tools proto-toolchain proto-format proto-lint proto-generate proto-check proto-breaking

build: proto-generate
	go build ./...
	go build -o bin/model ./model/cmd

test: proto-generate
	go test ./...
	cd model && go test ./...

race: proto-generate
	go test -race ./...
	cd model && go test -race ./...

vet: proto-generate
	go vet ./...
	cd model && go vet ./...

sync:
	go work sync

run: proto-generate
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

proto-check: proto-lint proto-generate
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
