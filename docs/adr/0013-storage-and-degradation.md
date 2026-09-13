# 13. Shared store with in-process read-through cache

Date: 2026-08-30
Status: Partially deferred

## Context

Hard fail (ADR-0002) requires high availability, which requires stateless
replicas. But the catalog and team policies are state. In-memory-only state
means replicas diverge — a policy applied to replica 2 is missing on replica 1 —
so more than one instance becomes impossible and HA is unachievable.

## Decision

State lives in a **shared store** (Postgres or etcd; choice deferred).

The service reads it through an **in-process read-through cache** with a short
TTL.

## Deferred

Which store, and the schema.

## Not deferred: behavior during a store outage

This is a design decision, not a storage detail, and determines whether the
store sits *on* the hot path or *beside* it.

**Decision needed:** serve inference from stale cache when the store is
unreachable, failing only on a cache miss — registration and `policy apply` go
read-only, inference keeps working.

Without this, the service inherits the store's availability and the HA that
hard-fail was chosen for is undone by a database restart.

## Consequences

- The store becomes a new single point of failure; the cache is what stops that
  mattering for inference.
- Discovery (`ListModels`) reads the store directly and **may** fail during a
  store outage. That is correct — no production traffic depends on a listing.
  See ADR-0006.
- Policy versions are stored, enabling instant rollback (ADR-0010).
