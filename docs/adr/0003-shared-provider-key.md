# 3. One shared provider key per provider

Date: 2026-08-30
Status: Accepted (known limitation)

## Context

The service holds provider credentials on behalf of all tenants. Keys can be
issued per-tenant (isolation) or shared across the org (utilization).

## Decision

**One shared key per provider** for v1. Per-tenant usage is *recorded* from day
one, but not *enforced*.

## Rationale

Per-tenant keys fragment one rate-limit pool into fixed slices: an idle team's
quota is stranded while a busy team is throttled next to it. Per-tenant
enforcement requires a distributed limiter and, more fundamentally, requires
tenant identity, which is not yet decided.

## Known limitation: noisy neighbor

All tenants share one rate limit (TPM/RPM) and one spend pool. There is no
per-tenant isolation.

**Consequence.** A single tenant issuing high-volume or long-context traffic —
batch backfills, load tests, runaway retry loops — can exhaust the org limit and
cause 429s for unrelated tenants, including user-facing traffic. Before this
service existed, teams held their own keys and this failure was self-contained;
centralizing converted it into a shared failure mode.

**Detection.** Alert on provider 429 rate. Per-tenant token and request counters
must be recorded from day one so the offending tenant is identifiable during an
incident.

**Mitigation today.** Manual: identify the tenant from metrics, ask them to stop.

**Planned fix, cheapest first.**
1. Split keys by *traffic class* (`interactive` / `batch`) so batch work cannot
   consume interactive capacity. Highest value per unit of work; needs no tenant
   identity.
2. Per-tenant token buckets.
3. Per-tenant provider keys.

**Blocked on.** Tenant identity — none of the above is implementable without it.

## Consequences

- Per-tenant request/token/cost counters are a v1 requirement, not a nice-to-have.
- Retry limits are enforced with platform-owned caps (ADR-0009) precisely
  because one team's retry configuration is everyone's problem.
