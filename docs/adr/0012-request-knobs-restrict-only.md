# 12. Request-level knobs may restrict behavior, never expand it

Date: 2026-08-30
Status: Accepted

Follow-up: [ADR-0031](0031-model-names-without-snapshots.md) supersedes the
family-alias wording below: disabling fallback restricts catalog model selection
without freezing the provider's underlying implementation.

## Context

Some behavior genuinely needs to vary per call — an eval job wants
reproducibility, a latency-sensitive path wants a tighter budget. But every
expressive field in the request proto is a value 40 teams may hardcode, and
therefore a migration the platform has pre-signed.

## Decision

> Request-level knobs may **restrict** behavior. They may never **expand** it.

Allowed in the request:

- `allow_fallback: false` — this model or fail. Required for eval and
  reproducibility work.
- `max_attempts`, deadline, budget ceiling — all bounded by platform caps.
- `require_capabilities` — fail rather than fall through to a target lacking
  tool support.

Not allowed in the request:

- `fallback_chain: [...]` or any other expression of routing policy.

## Rationale

A restriction is scoped to one call, creates no configuration surface to
maintain, and cannot surprise the platform later — worst case, one team fails
closed and only they are affected.

An expansion is permanent, org-wide, and unfixable from the platform side once
it is hardcoded.

The design test for any new proto field:

> *If 40 teams hardcode a value for this, can the platform still do its job?*

If no, it belongs in the catalog or in team policy.

## Consequences

- `allow_fallback: false` makes ADR-0004's family-alias behavior opt-out per
  call, which is what reproducibility-sensitive work needs.
- Adding an expanding field later is easy; removing one is not. Bias toward
  omission.
