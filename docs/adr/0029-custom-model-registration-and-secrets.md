# 29. Organization-wide custom models with reviewed endpoints and secret references

Date: 2026-11-01
Status: Accepted (registration schema and secret backend partially deferred)

## Context

Teams need to register custom and self-hosted endpoints for organization-wide use.
Names share the flat catalog in ADR-0004. Registration must not silently redirect
existing consumers or expose endpoint credentials.

## Decision

### Registration, visibility, and updates

- Support custom/self-hosted model registration through an authenticated service
  API. Once approved and active, entries are available to all organization tenants;
  v1 has no tenant-only model visibility. Provider allowlists, capability checks,
  and other tenant authorization still apply.
- No-price registrations requesting `unmetered` remain pending and unusable until
  platform approval as specified in ADR-0028.
- Identical model registration returns the existing entry. Reusing an existing
  name with conflicting settings is rejected rather than treated as an update.
- Changes require an explicit update authorized for the owning tenant or the
  platform team. Registration cannot silently replace another team's model.
- Endpoint changes require renewed platform approval. Keep the current approved
  endpoint active while the replacement is pending. Switch to the replacement
  only on approval; rejection leaves the current endpoint unchanged.
- Credential rotation under existing ownership permissions does not require
  endpoint reapproval and requires no consumer application redeploy.

### Credential lifecycle

- Teams may submit an endpoint credential in an authenticated, encrypted
  registration request. Treat that field as write-only secret input.
- Store the credential in a platform-managed secret manager. HashiCorp Vault is a
  candidate, not a separately finalized backend/vendor choice.
- PostgreSQL stores the model configuration and secret reference, not plaintext
  endpoint credentials. The model service resolves the reference for provider
  calls. Other tenant applications do not receive those credentials.
- Never expose secret values in registration responses, discovery/catalog queries,
  logs, traces, or errors. ADR-0022's diagnostic content restrictions apply.
- Update the managed credential on rotation without requiring model re-registration
  or consumer application redeployment. Runtime credential caching must permit
  rotation; exact refresh mechanics remain to be designed.

### Catalog write-path follow-up

The runtime registration/update API adds a catalog write path to ADR-0024's
reviewed-config/CI default. Platform-managed baseline entries may still use that
path; team custom registrations use the authenticated service path and PostgreSQL.
CI must not silently overwrite runtime-owned custom entries or pending updates.
Specify ownership and reconciliation between these paths before implementation.

## Consequences and remaining work

One custom endpoint may now serve multiple teams. The platform approval process
must consider organization-wide reachability/capacity. Endpoint staging preserves
availability during review, while global names and explicit updates protect
existing callers from accidental redirection.

Specify the model/approval schema, equality rules (including secret inputs), model
capability validation, allowed endpoint destinations, secret-manager permissions,
partial registration/secret-write cleanup, credential rotation refresh, and
approval audit/concurrency controls. Do not infer broader paid-model approval or
automatic activation rules beyond the accepted no-price and endpoint-change gates.
