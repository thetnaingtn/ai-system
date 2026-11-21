# 33. Platform approval for all new model registrations

Date: 2026-11-15
Status: Accepted

## Context

Approved models are available organization-wide, regardless of the owning tenant.
ADR-0028 requires platform approval for unmetered models, and ADR-0029 requires
renewed approval for endpoint changes but leaves broader priced-model activation
rules open. The owner chose to require approval for metered registrations too.

## Decision

- Every new model registration, metered or unmetered, requires platform approval
  before it can serve generation requests. This applies to both team-owned and
  platform-owned registrations; ownership alone does not bypass approval.
- New registrations start with `approval_status = pending`. Only an authorized
  platform approval action may transition a registration to `approved`.
- A rejected registration cannot serve. Retrying an identical registration is not
  a route around a pending or rejected decision.
- Approval does not bypass pricing, authorization, capability checks, lifecycle
  restrictions, or tenant limits. Metered models still need applicable pricing.
- Preserve ADR-0029's requirement to stage endpoint changes for renewed approval
  while continuing to serve the current approved configuration. Credential
  rotation under existing permissions does not require endpoint reapproval.

## Supersession

This resolves ADR-0029's open approval/activation policy for priced registrations
by extending platform approval to all new registrations. It retains ADR-0028's
unmetered approval requirements and the other ADR-0029 ownership, update, and
secret-management decisions. Historical text is preserved.

## Consequences

Registration and activation are separate for every model. Registration responses
must communicate pending approval; the approval workflow must be implemented
before newly registered models can serve. This decision does not require a
separate human reviewer from the registrant; review identity and audit mechanics
remain implementation details.
