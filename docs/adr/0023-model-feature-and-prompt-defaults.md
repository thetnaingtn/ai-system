# 23. V1 model features and remaining prompt defaults

Date: 2026-11-01
Status: Accepted

## Context

The user delegated technical defaults after settling prompt ownership, immutable
revisions, release selection, and per-call fallback controls. This records those
choices and resolves the unratified items in ADR-0015 and ADR-0016.

## Decision

- Defer structured-output emulation beyond v1. Effective structured-output support
  in v1 is native or none; remove unsupported targets from eligible routing rather
  than weakening the output contract. ADR-0015's emulation architecture remains
  available for future implementation.
- Reasoning effort remains provider-specific passthrough in v1, subject to existing
  vendor-parameter validation. Do not introduce a normalized reasoning enum yet.
- Ratify `pin` as the default tool-conversation failover policy once tool results
  exist. Retain ADR-0016's explicit tenant-policy opt-out for teams accepting that
  risk. This does not provide idempotency for application-owned side effects.
- If a provider violates an explicitly requested parallel-tool restriction, return
  a typed contract error before exposing executable tool calls. Do not silently
  accept the violation. Streaming validation/buffering must preserve this guarantee;
  partial tool argument events must not be treated as approved executable calls.

### Prompt lifecycle follow-up

- The first release is explicit. Registration does not create or move `stable`,
  including on initial registration. Default selection fails clearly until the
  first release. The release protocol must distinguish an expected absent tag
  from a tag already created by another publisher.
- Prompt scope selection is explicit. Do not silently resolve a missing tenant
  prompt to an organization prompt of the same name.
- Preserve referenced revisions. Defer physical prompt deletion in v1; reference
  tracking and later deletion rules require a separate design.

## Consequences

This settles the four model-feature questions carried from ADR-0015/0016 and the
first-release, scope-lookup, and v1 deletion defaults from ADR-0018. Native-only
structured output reduces fallback coverage. Tool contract validation can require
buffering before returning approved calls. Neither choice implies prompt quality
equivalence across models.
