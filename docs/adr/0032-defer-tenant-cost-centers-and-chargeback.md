# 32. Defer tenant cost centers and internal chargeback

Date: 2026-11-08
Status: Accepted

## Context

The owner excluded internal chargeback from the initial model service and agreed
to omit the tenant cost-center field. ADR-0014 previously required both an owner
and cost center before a tenant could leave the experimental tier.

## Decision

- Omit `cost_center` from the initial `tenant_config.tenants` table.
- Production tenants still require a responsible owner contact. A cost center is
  not required for production access.
- Defer internal chargeback and billing teams. Revisit cost-center metadata only
  when a concrete attribution or chargeback requirement needs it.
- Preserve authenticated tenant identity, usage tracking, rate/concurrency limits,
  and monthly provider-spend budget enforcement. Metered/unmetered classification
  and ADR-0028's reservation/reconciliation requirements remain in effect.

## Supersession

This supersedes ADR-0014's production cost-center requirement and initial
chargeback scope. Its tenant authentication, experimental defaults, ownership,
and authorization decisions remain accepted. Historical text is preserved.

## Consequences

Tenant metadata is smaller and does not require billing metadata for promotion
to production. Provider-spend amounts enforce budgets and support usage reporting;
they do not create invoices or internal transfers between teams.
