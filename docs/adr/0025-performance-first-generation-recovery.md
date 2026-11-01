# 25. Generation recovery favors performance over strict duplicate prevention

Date: 2026-11-01
Status: Accepted (recovery mechanics partially deferred)

## Context

ADR-0019 and ADR-0024 prohibited automatic regeneration when a provider outcome
was uncertain. The owner clarified that occasional duplicate model generation is
acceptable: its cost is additional provider spend and potentially different text.
Avoid adding stronger coordination or durable execution records solely to prevent
that duplication.

## Decision

- Keep Redis atomic claims, request fingerprints, owner leases, and completed
  response caching as best-effort coordination. Healthy replicas still join an
  existing running operation or return its cached response.
- If an owner lease expires, a result cannot be recovered, or Redis loses an
  operation record, recovery may claim the operation and generate again. This
  supersedes the prohibition on automatic regeneration in ADR-0019 and ADR-0024.
- Do not require a PostgreSQL execution ledger or introduce a consensus/strong
  consistency layer solely to prevent duplicate generation. Redis data loss or
  failover may cause duplicate work; there is no exactly-once execution promise.
- Regeneration remains bounded by the original deadline and existing attempt and
  retry policies; it does not grant a new SDK retry allowance or a fresh deadline.
  An expired lease does not establish that the old provider request has stopped,
  so overlapping provider work is possible during recovery.
- Keep completed responses for 10 minutes. Reuse the same SDK-generated key for
  retries, and reject changed request content when an existing fingerprint is
  available. Lost records also lose the corresponding validation evidence.
- Redis unavailability before starting a protected request still fails closed.
  Tolerating loss of a record after Redis recovers does not authorize executing
  without Redis coordination while Redis is unreachable.
- Continue honoring cancellation and completed states that are available. Do not
  knowingly regenerate a cancelled operation. Redis record loss can also lose
  cancellation evidence; this limits the cancellation guarantee during failure.
- Prevent an obsolete owner's late completion from overwriting a replacement
  owner's state, using ownership/attempt tokens checked in atomic transitions.
- Record recovery/takeover attempts and observable token/cost usage. Exact
  duplicate provider cost may be unknowable when the original request's result
  and usage are lost; do not report unobserved cost as zero.

## Scope and consequences

This tolerance applies to model generation, including generation that returns
tool-call proposals. It does not authorize automatic replay of application-owned
tools, payments, product creation, or other side effects. Their APIs need their
own action-level idempotency and business constraints.

Recovery can produce a different answer or tool-call proposal, and an original
provider request may continue billing alongside its replacement. More than one
extra charge is possible across repeated failures; duplication is bounded where
existing deadlines, attempts, and admission controls can be enforced, not by an
exactly-once guarantee. Streaming retains ADR-0008's no-replay-after-output rule.

An operation still returns an error when recovery cannot finish within its
remaining budget. This decision permits regeneration; it does not guarantee
recovery of an answer after every failure or add a result-retrieval API.

## Deferred

Lease durations, takeover scripts, original-budget propagation, cancellation/
completion races, and Redis failover configuration remain implementation work.
Exact cross-replica attempt accounting after record loss is necessarily limited
without durable evidence; avoid presenting best-effort coordination as a strict
global execution bound.
