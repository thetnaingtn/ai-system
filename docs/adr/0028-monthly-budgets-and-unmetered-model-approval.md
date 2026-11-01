# 28. Monthly spend caps and approved unmetered models

Date: 2026-11-01
Status: Accepted (accounting details partially deferred)

## Decision

- Tenant provider-spend budgets reset at 00:00 UTC on the first day of each month.
- Reject new generations once the applicable monthly budget is exhausted. Allow
  already admitted in-flight requests to finish. Raising a budget is an authorized
  policy change, not a per-request bypass.
- Reserve an estimated maximum provider cost before admission and reconcile with
  actual usage after execution, so concurrent requests cannot all spend the same
  available budget. Preserve separate rate and concurrency limits for short-term
  shared-capacity protection.
- A paid provider model with missing pricing is blocked: unknown cost is not zero.
- Self-hosted/custom models may request an explicit `unmetered` designation rather
  than supply artificial token pricing. This means excluded from provider-spend
  accounting, not free infrastructure or unlimited capacity.
- The registering team can request that designation, but the platform team must
  approve it. A registration without pricing remains pending and unusable until
  approval. Do not allow unapproved no-price use during review.
- Approved unmetered models require no token pricing and are not charged against
  monthly provider-spend budgets. Usage tracking, rate/concurrency limits, tenant
  authorization, and technical capability checks still apply.
- A custom endpoint backed by a paid provider is not exempt merely because it is
  custom. Its classification and any exemption must be reviewed appropriately.

## Consequences and remaining work

This resolves the budget reset/action, unknown-price policy, and no-price approval
choices left in ADR-0024. A tenant cannot bypass paid usage accounting simply by
marking an endpoint unmetered.

Specify numeric quotas, reservation estimates and storage, retry/recovery cost
reservations, missing-usage reconciliation, budget-period attribution across month
boundaries, and enforcement-state outage behavior. Actual duplicate provider costs
can be uncertain under ADR-0025; do not promise exact global spend enforcement
without resolving how uncertain costs are reserved/accounted for.
