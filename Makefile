MIGRATE_VERSION := v4.20.1
MIGRATE := $(CURDIR)/bin/migrate
MODEL_DATABASE_URL ?= postgres://model:model-local@127.0.0.1:5432/model?sslmode=disable&search_path=public

.PHONY: build test race vet test-migrations sync run tools migrate-up migrate-down migrate-create

build:
	go build -o bin/model ./model/cmd

test:
	cd model && go test ./...

race:
	cd model && go test -race ./...

vet:
	cd model && go vet ./...

test-migrations:
	cd model && MODEL_MIGRATION_TEST=1 go test ./migrations -run TestMigrations -count=1 -timeout=4m -v

sync:
	go work sync

run:
	go run ./model/cmd

tools:
	GOBIN="$(CURDIR)/bin" go install -tags postgres -ldflags="-X main.Version=$(MIGRATE_VERSION)" github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)

migrate-up:
	@"$(MIGRATE)" -path model/migrations -database "$(MODEL_DATABASE_URL)" up

migrate-down:
	@"$(MIGRATE)" -path model/migrations -database "$(MODEL_DATABASE_URL)" down 1

migrate-create:
	@test -n "$(name)" || (echo 'usage: make migrate-create name=describe_change' >&2; exit 1)
	"$(MIGRATE)" create -ext sql -dir model/migrations -seq "$(name)"
