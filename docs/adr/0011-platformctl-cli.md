# 11. A thin CLI is the build-time interface for humans and CI

Date: 2026-08-30
Status: Accepted

## Context

Two things need a non-application interface to the platform: developers
discovering what models exist, and CI validating and publishing team policy.

## Decision

Ship `platformctl`, a **thin client over the same gRPC API**. No separate
backend, no separate auth, no logic of its own.

```
platformctl models list                    -> ListModels
platformctl models get gpt-4o              -> GetModel
platformctl policy validate ./llm.yaml     -> ValidatePolicy (dry run)
platformctl policy apply    ./llm.yaml     -> ApplyPolicy
platformctl policy diff     team-a         -> GetPolicy + local diff
platformctl policy rollback team-a         -> RollbackPolicy
platformctl usage --model gpt-4-turbo      -> GetUsage
```

If business logic starts appearing in the CLI, it belongs in the service.

## Rationale

CI needs a scriptable binary regardless, so ship that before any web UI. The CLI
is the build-time surface from ADR-0006: a developer runs `models list` while
writing code, their CI runs `policy apply` on merge.

## The rule that keeps it from becoming a liability

**The CLI must not reimplement validation.** If CI validates with one
implementation and the server with another, the failure is a green build
followed by a rejected apply, and nobody knows which is right.

Either the CLI calls a `ValidatePolicy` RPC (preferred — one implementation), or
validation lives in a shared Go package compiled into both (faster local
feedback, requires pinning the CLI version in CI).

## Consequences

- A third artifact to release and distribute. A single Go binary keeps this
  cheap, provided the CLI stays thin.
- It must use the same tenant identity mechanism as the SDK, not a second one.
- **Minimum v1:** `models list`, `policy validate`, `policy apply`. Everything
  else is deferrable — none of it is on the request path.
