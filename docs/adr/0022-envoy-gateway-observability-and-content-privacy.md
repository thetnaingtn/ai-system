# 22. Envoy Gateway and automatic observability without content logging

Date: 2026-11-01
Status: Accepted

## Context

Teams should call a stable service endpoint without configuring backend discovery
or manually creating tracing spans. The user selected Envoy Gateway and prohibited
diagnostic logging of prompt and response contents.

## Decision

- Use Envoy Gateway on Kubernetes as the gRPC-aware API gateway. Clients address
  one stable endpoint; gateway configuration routes and balances individual RPCs
  to model, data, and tooling backends. Do not require application-owned
  client-side backend discovery/balancing.
- Run multiple gateway replicas. Support gRPC streaming, forwarded authentication
  metadata, and operation-aligned deadlines. Service-side authenticated identity
  checks remain in place; this does not introduce trust in arbitrary tenant headers.
- Disable gateway generation retries. ADR-0009 and ADR-0019/0020 retain retry
  ownership in the SDK and model service.
- The gateway continues valid incoming W3C trace context or starts a new trace
  when absent. Propagate standard trace context, not a proprietary trace-ID scheme.
- Platform-owned service/provider instrumentation supplies spans, metrics, and
  correlated operational logs automatically. Platform SDK wrappers propagate
  context without teams manually creating spans for ordinary platform calls.
- Teams may add custom application observability. Bare generated clients still
  receive gateway-to-provider visibility; joining earlier application work needs
  context propagation through a wrapper or automatic client instrumentation.
- Trace identity and idempotency identity are separate and serve different purposes.

### Diagnostic content policy

- V1 prohibits prompt, document, conversation, memory, tool payload, and response
  bodies in diagnostic logs and trace attributes/events. Do not add an opt-in body
  logging path in v1.
- Sanitize provider error details and exception paths so errors do not leak these
  contents. Credentials and authorization headers must not be logged either.
- Record operational metadata such as tenant, model/snapshot, prompt revision,
  latency, token counts, safe error categories, and trace identifiers.
- Enforce the same policy in gateway access logs, service logs, provider adapters,
  telemetry exporters, and diagnostic hooks; avoid uncontrolled metadata capture.
- The Redis response-recovery cache and deliberate prompt/session/memory storage
  are functional storage, not diagnostic logging. This policy does not remove
  them or establish their complete access/retention contracts.

## Consequences

This settles Q14 and observability/logging policy, replacing the earlier delegated
client-side round-robin recommendation with the user's gateway choice. Envoy
Gateway is an additional runtime dependency. A gateway trace ID alone cannot
describe internal service work; automatic service instrumentation remains required.
Full historical application traces cannot be inferred from an uninstrumented client.

Gateway installation, route policies, sampling/exporter settings, metadata
allowlists, and telemetry retention still require deployment specifications.
