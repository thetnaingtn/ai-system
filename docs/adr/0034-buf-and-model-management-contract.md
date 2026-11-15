# 34. Buf tooling and the initial model-management protobuf contract

Date: 2026-11-15
Status: Accepted

## Decision

- Keep model sources at `proto/platform/model/v1/`, with protobuf package
  `platform.model.v1` and service `ModelService`.
- Use root Buf v2 configuration: `buf.yaml` defines the `proto/` module with
  `STANDARD` lint rules and `FILE` compatibility rules. `buf.gen.yaml` defines
  Go message and gRPC code generation.
- Use the separately installed Buf CLI, initially version 1.75.0. Do not install
  another Buf binary into the repository. The Makefile checks the expected version.
- Manage Go generation plugins through Go `tool` directives and versioned module
  requirements: protoc-gen-go v1.36.11 and protoc-gen-go-grpc v1.6.2 initially.
  Buf invokes `go tool` locally; no standalone plugin install, remote generator,
  or BSR publication is needed.
- Commit generated Go files under `gen/go/platform/model/v1/` in the root module.
  The model module remains independently buildable until it starts importing
  these shared contracts; its image does not need changes before that import.
- Define registration, discovery, updates, change review, and credential rotation
  first. Generation, prompts, policies, and ergonomic Go SDK wrappers follow.
- Keep public model metadata separate from authorized configuration views and
  change proposals. Credential inputs are write-only, and secret-manager references
  never appear in the contract. Tenant identity is resolved at the transport boundary.
- Require expected configuration versions for updates, approvals, and credential
  rotation. Stage endpoint/provider and accounting-mode changes for platform review;
  a staged request applies its complete mutation together, without partial updates.
- Use separate request/response messages for each RPC, typed capability messages,
  explicit enum zero values, paginated lists, mutation request IDs, and typed gRPC
  error details. Detailed behavior is documented in the model API design.
- CI runs only on pull requests, checking formatting, lint, generation freshness,
  compilation, and compatibility against the PR base. The first change introducing
  `buf.yaml` establishes the compatibility baseline. Future changes must pass
  compatibility checks; intentional breaks require a new API version.

## Consequences

Go module tools share dependency selection with the module/workspace. Review
tool/runtime upgrades together and regenerate after intentional version changes.
Buf lint and compatibility checks do not enforce runtime authorization, business
validation, or semantic compatibility; service code must implement those rules.

The generated service interfaces are contracts only. The running model process
still exposes only health checks. Pending-change persistence, safe write replay,
secret-manager integration, and authenticated handlers must be implemented before
these management RPCs are registered. The initial model-change table design remains
a draft; defining its API does not create its storage or implementation.

## References

- [Model API design](../ai-platform-model-api.md)
- [Buf generation configuration](https://buf.build/docs/configuration/v2/buf-gen-yaml/)
- [ADR-0021](0021-service-boundaries-and-sdk-contracts.md)
- [ADR-0033](0033-approval-for-all-model-registrations.md)
