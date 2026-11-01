# 26. PostgreSQL for durable data, Redis for temporary state; model service first

Date: 2026-11-01
Status: Accepted

Follow-up: [ADR-0027](0027-prompt-equality-and-postgres-degradation.md) settles
exact prompt equality and cached inference during PostgreSQL outages. The original
remaining-work list below is retained as history; detailed schemas and cache
implementation still require specification.

## Decision

- Use PostgreSQL for durable platform records, including the model catalog,
  tenant policies, and prompt revisions/releases. This settles ADR-0013's
  PostgreSQL-versus-etcd storage choice.
- Use Redis for temporary state such as generation coordination and cached
  results. Its existing 10-minute completed-response retention remains unchanged.
- Build the first model service before designing the data/session/memory services
  in detail. Their boundaries remain as recorded in ADR-0021; this is sequencing,
  not removal from the overall platform.
- Provide the platform SDK in Go for now. Protobuf/gRPC remains the service
  contract, so authorized teams can generate clients in other languages. Additional
  platform-maintained language wrappers are deferred.

## Remaining work

PostgreSQL schema, migrations, and transaction boundaries remain to be specified.
ADR-0013's stale-cache/store-outage policy, cache misses, and cold-start behavior
are not settled merely by choosing PostgreSQL. Prompt content equality is also
still open; exact-text comparison was proposed but has not been accepted.
