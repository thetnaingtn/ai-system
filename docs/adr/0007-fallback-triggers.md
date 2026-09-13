# 7. Which failures trigger fallback, and which never do

Date: 2026-08-30
Status: Accepted

## Context

"Fall back on failure" is underspecified. Different failures have different
correct responses, and two of them are policy decisions rather than engineering
ones.

## Decision

| Failure | Retry same target | Fall back | Note |
|---|---|---|---|
| 429 rate limit | yes, once, honouring `Retry-After` | yes | immediate cross-provider failover on 429 stampedes the fallback |
| 5xx / connection error | yes | yes | the clean case |
| Timeout | yes, with idempotency key | yes | request may have succeeded server-side |
| `context_length_exceeded` | no | **no** | deterministic; recurs unless the chain is capability-validated. Fail fast with a clear error |
| Content-filter refusal | no | **no** | a valid model response, not a failure |
| 401 / 403 | no | **no** | platform configuration bug; fail and page |

A non-retryable error breaks out of the retry loop for that target but does
**not** abort the chain, except for the three marked "no" above, which terminate
the request.

## Rationale

- **Content filtering.** Routing around another provider's safety control would
  make the platform a safety bypass, and the caller needs to see the refusal.
  This is a deliberate policy the platform team should be able to state.
- **Context length.** Falling back just fails again on a smaller window. Because
  chains are capability-validated at config time (ADR-0005), a coherent chain
  cannot fix this by trying another hop.
- **401/403.** Falling back would mask a broken or rotated key and silently
  migrate all org traffic to a different provider at different cost and quality.
  Loud failure is correct, especially during rollout.

## Consequences

- The error taxonomy is part of the public contract; teams will branch on it.
  Typed errors are required (see open questions).
- Every terminal error must report per-target detail: which targets were tried,
  what each returned, and why the chain stopped.
