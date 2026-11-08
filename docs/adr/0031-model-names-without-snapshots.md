# 31. Model names without a separate snapshot concept

Date: 2026-11-08
Status: Accepted

## Context

The owner chose model-name-based selection, such as `gpt-4o`, without a separate
snapshot abstraction. All clients use the same real model name; different aliases
for the same model would create confusion. This reinforces ADR-0004's rejection
of invented intent names such as `summarize-cheap`. The model service schema and
application APIs are still being designed; the proposed snapshot layer has not
been implemented.

## Decision

- Use one globally unique model `name`, such as `gpt-4o`, consistently across
  all clients. Retain ADR-0004's flat, organization-wide namespace, lifecycle
  status, capabilities, and pricing.
- The registered model name is also the model identifier sent to its provider
  endpoint. Store the provider and endpoint alongside it; do not introduce a
  separate catalog alias or independently configurable provider model ID.
- Do not register alternative client-specific or intent-based names for the
  same model. Applications may choose different models for their tasks, but
  model selection uses the shared real model names.
- Do not introduce snapshot tables, snapshot fields, family-to-snapshot routing,
  snapshot selection options, or snapshot-pinning SDK methods.
- Resolution responses report the requested model name, actual selected model
  name, provider, and catalog configuration version,
  alongside the existing attempt, warning, and degradation information. Do not
  require an underlying model version that the endpoint does not expose.
- Configuration versions describe changes to platform-managed routing and model
  configuration. They do not identify or guarantee immutable model weights.
  The exact configuration history schema remains implementation work.
- Disabling fallback restricts a request to its selected model; it does
  not freeze the implementation behind that name. The platform makes
  no model reproducibility guarantee based on a model name.
- Preserve immutable prompt revisions, prompt release semantics, and the
  tool-conversation model-pinning policy. These do not require model snapshots.
  Unary retries retain the original resolved model configuration while its
  operation record is available, as required by ADR-0020.

## Supersession

This supersedes the snapshot mapping, selection, rotation, and reporting
requirements in ADR-0004; snapshot reporting in ADR-0006; and the snapshot-based
evaluation guidance in ADR-0018. It also supersedes the snapshot/version-pinning
wording in ADR-0001, the snapshot example in ADR-0015, the family-alias wording in
ADR-0012, and snapshot observability metadata in ADR-0022. Their unrelated
decisions remain in effect. Historical text and examples are preserved.

## Consequences

The catalog and generation API use one model name without an alias mapping.
Endpoint implementations may change behind the same model name, so configuration
history alone cannot explain every output change. Endpoint approval, credential
rotation, pricing, capability validation, authorization, and fallback policies
remain required under their existing ADRs. Detailed schema and API choices remain
open for discussion.
