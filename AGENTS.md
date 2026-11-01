# Repository Guidelines

## Project Structure & Module Organization

This repository is currently an architecture-first Go module:
`github.com/thetnaingtn/ai-system`.

- `go.mod` declares Go 1.26.3.
- `docs/adr/` contains numbered architecture decision records and their index.
- `docs/adr/open-questions.md` tracks remaining model-service implementation work.
- `.agents/skills/` contains repository-local agent skills.

Application packages, tests, deployment manifests, and generated protobuf code
have not been added yet. Establish their layout explicitly during implementation;
do not assume existing `cmd/`, `internal/`, or `proto/` directories.

## Architecture & Decision Records

Implement the model service and Go SDK first. Use protobuf/gRPC, PostgreSQL for
durable records, Redis for temporary coordination/results, and Envoy Gateway.
Data, sessions, memory, and workflow implementation follow later.

Read relevant ADRs and their follow-up links before changing behavior. Later
records supersede earlier decisions where stated. Add sequential records such as
`0030-descriptive-decision.md`, update the ADR index, and preserve decision history.

## Build, Test, and Development Commands

Run from the repository root once Go packages exist:

- `go build ./...` — compile all packages.
- `go test ./...` — run the test suite.
- `go test -race ./...` — check concurrent code for data races.
- `go vet ./...` — run standard static checks.
- `gofmt -w path/to/file.go` — format changed Go files.
- `git diff --check` — catch whitespace errors in tracked changes.

There is currently no runnable service, Makefile, or configured lint pipeline.
Document startup, migration, and protobuf-generation commands when introduced.

## Coding Style & Naming Conventions

Use standard Go formatting with tabs via `gofmt`, lowercase package names,
`MixedCaps` identifiers, and descriptive filenames. Keep generated protobuf types;
use small wrappers for SDK errors, streaming, and platform behavior. Keep tenant
identity at the authenticated transport boundary.

## Testing Guidelines

Use Go's `testing` package, `*_test.go` files, and `TestBehavior` names. Prefer fake
providers and injectable clients over live paid calls. Cover cancellation,
concurrent registration, retries, and budget accounting. No coverage threshold is
currently defined.

## Commit & Pull Request Guidelines

History favors `docs: ...` and `chore(scope): ...`; use concise imperative subjects.
Keep commits scoped. PRs should explain the behavior, relevant ADRs/issues,
validation performed, and material limitations. State when checks cannot run.

## Security & Configuration

Never log prompt/response bodies, tool payloads, or credentials. Store endpoint
credentials in a secret manager and only references in PostgreSQL. Do not commit
secrets or weaken tenant authorization for local convenience.
