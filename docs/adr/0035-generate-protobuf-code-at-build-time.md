# 35. Generate protobuf code at build time

Date: 2026-11-21
Status: Accepted

## Decision

- Exclude root `gen/` from source control. Track protobuf sources, Buf
  configuration, and the Go manifests/checksums that select generator versions.
- Generate Go code under `gen/go/` before local builds, tests, vet, and startup
  through the existing Make targets. Direct Go commands require generation first.
- Keep CI pull-request-only. Generate before compiling instead of comparing
  generated files with committed copies. Continue protobuf lint, formatting, and
  source-schema compatibility checks.
- Exclude local `gen/` from the Docker build context and generate inside the image
  build with pinned Buf and Go module tools. A fresh checkout must not depend on
  developer-generated files being present.

## Supersession

This supersedes ADR-0034's decision to commit generated Go files and check their
freshness against Git. Its source layout, generator pins, local `go tool`
invocation, API contracts, and schema compatibility rules remain accepted.

## Consequences

Generated code remains available locally after generation but does not appear in
Git changes. Fresh checkouts and consumers of this source tree must run generation
before compiling packages that import the generated types. Distribution of a
pre-generated Go SDK is separate future work.
