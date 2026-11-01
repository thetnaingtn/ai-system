# 24. V1 operating policy, tenant protection, and recovery defaults

Date: 2026-11-01
Status: Accepted (implementation values and storage details deferred)

## Context

The user accepted tenant limits and production protection, selected best-effort
availability for v1, and delegated remaining operational defaults. Earlier records
assumed a 24/7 rotation and accepted unenforced noisy-neighbor risk; these parts
need explicit supersession rather than treating them as current requirements.

## Decision

### Availability and admission

- V1 offers best-effort availability with no contractual availability promise or
  24/7 support commitment. Communicate this to adopting teams. This supersedes
  ADR-0002's mandatory 24/7 on-call and published-SLO commitment for v1.
- Keep ADR-0002's hard-fail/no-client-bypass rule and credential centralization.
  Best effort does not introduce a provider bypass or remove basic resilience.
- Enforce tenant limits, experimental hard caps, production capacity protection,
  and configurable production spend ceilings in v1. This supersedes ADR-0003's
  per-tenant recording-without-enforcement decision while retaining shared provider
  keys. Tenant identity from ADR-0014 supplies the enforcement key.
- Exact limits, accounting/reservation rules, priority allocation, and behavior
  when limit/accounting state is unavailable require implementation specifications.

### Deployment and control plane

- On shutdown, stop admitting new generation work and drain existing streams and
  operations, including unary operations whose waiting RPC disconnected. Bound
  draining by operation deadlines and allow a shutdown margin in infrastructure.
- Use one original operation deadline across retries; fit provider attempts inside
  the remaining budget. Infrastructure request/idle timeouts must support the
  longest allowed operation plus margin. Do not give retries a fresh deadline.
- The platform team owns pricing metadata through reviewed updates. Missing prices
  mean unknown cost, never zero. Cost-cap behavior for unknown prices is still open.
- Author the platform catalog as reviewed repository configuration, validated and
  published by CI. Team policies keep ADR-0010's authoring/distribution contract.
- Provide a narrow authorized, audited incident operation to disable a target,
  with a 24-hour expiry. Keep the override distinct from baseline configuration so
  it cannot become silent permanent drift.

### Unary recovery implementation defaults

- Use atomic Redis operation claims/completion transitions, request fingerprints,
  and bounded owner leases. Share completion and cancellation state across replicas.
- A retry waits for the original owner/result within existing deadlines. Owner loss
  must not automatically restart an operation whose provider outcome is uncertain;
  retain the outcome-unknown behavior from ADR-0019.
- Cancellation must reach or be observed by the executing owner, not merely be
  written as an unenforced flag. Preserve completed results on repeated cancellation.
- Lease durations, fingerprint normalization, cross-replica notifications/waiting,
  cancellation/completion races, Redis failover/record-loss behavior, and provider
  timeout ambiguity remain detailed design work. These defaults are not a proof of
  exactly-once execution or a complete Redis coordination protocol.

## Consequences

This resolves the policy choices for Q13, Q16, Q25, catalog writes, incident
overrides, budgets, and v1 support expectations. Tenant protection adds launch
work compared with ADR-0003. Best-effort availability must not be represented as
a guaranteed production SLO. ADR-0013's durable store selection, schema, and cache
outage semantics remain undecided; Redis recovery records do not choose that store.
