# 2. Service unavailability is a hard fail; clients never bypass to providers

Date: 2026-08-30
Status: Accepted

Follow-up: [ADR-0024](0024-v1-operating-policy-and-recovery-defaults.md). supersedes the mandatory 24/7 on-call and SLO commitment for v1 with best-effort availability. The hard-fail/no-bypass decision remains in force.

## Context

Centralizing LLM calls introduces a single point of failure that did not exist
when teams called providers directly. When the model service is down, something
has to happen.

Options considered:

1. **Hard fail** — calls error; every AI feature in the org degrades.
2. **Client-side bypass** — the SDK calls providers directly when the service is
   unreachable.
3. Cached or stale responses — meaningless for generative calls.

## Decision

**Hard fail.** The SDK does not hold provider credentials and cannot bypass the
service under any condition.

## Rationale

Option 2 requires the SDK to hold provider API keys, which destroys the primary
reason the service exists (ADR-0001). A bypass path is also the least-tested
code in the system and would run only during incidents.

## Consequences

- The model service must be highly available: multiple stateless replicas, no
  single-instance state. See ADR-0013.
- The platform team takes on a 24/7 on-call rotation and must publish an SLO.
- Deploys must drain in-flight streams gracefully rather than severing them.
- Availability of the service is now a company-level concern, and its outage
  budget should be treated as such.
