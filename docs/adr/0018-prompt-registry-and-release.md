# 18. Immutable prompt revisions with separate, authorized releases

Date: 2026-10-04
Status: Accepted

Follow-up: [ADR-0023](0023-model-feature-and-prompt-defaults.md). settles explicit first release, explicit scope lookup, and deferral of physical deletion. Content equality and detailed concurrency contracts remain implementation work.

## Context

Teams need centrally stored system prompts without requiring an application
redeploy for each prompt release. Registration must be usable from application
code, including an admin-panel backend, rather than restricted to CI.

## Decision

### Revisions and release selection

- The platform assigns immutable revisions within a prompt's scope and name.
  Teams do not supply competing version formats such as `0.1.0` or `version-1`.
- Changed content under an existing name creates a new revision. A new name
  starts a separate revision history.
- Registering identical content under the same scope and name returns the
  existing revision, including when registration is retried.
- Registration and release are separate operations. Registration does not move
  a release tag.
- Callers may select an exact revision. Otherwise selection defaults to the
  movable `stable` tag.
- A release API moves a tag to a selected revision; rollback moves the tag to
  an earlier revision. Moving a tag does not alter immutable content or pinned
  callers. An in-flight request keeps its already resolved revision.
- Responses report the actual prompt revision used.
- Releases require the expected current revision of the tag. If it has changed,
  reject the release with a conflict rather than overwriting another release.

### Scope and authorization

- Tenant scope is the default. Tenants also may use organization-scoped prompts.
  Scopes have separate revision histories and release tags.
- The platform team registers and releases organization prompts. Other tenants
  can consume them and pin any available exact revision.
- Tenant prompt writes require explicit workload authorization, using proven
  identity rather than a caller-supplied tenant claim.
- Registration and release are separately granted permissions
  (`prompt.register` and `prompt.release`). A workload may hold either or both.
- Authorized applications, including admin backends, can use the same
  `RegisterPrompt()` and release APIs as publishing automation.

### Prompt contents and generation

- V1 stores plain system prompt text. Teams construct changing user messages and
  conversation history. The platform performs no variable substitution in v1;
  template support is deferred to a later version.
- A generation request uses either a registered system prompt or an inline
  system message. Reject requests containing both.
- Teams own prompt/model quality compatibility and evaluations. V1 stores no
  prompt-specific model compatibility list. A caller can disable fallback per
  request using ADR-0012; the tenant policy otherwise governs fallback.
- Platform checks for technical requirements, such as tools and structured
  output, still apply. Disabling fallback does not pin a family alias's snapshot:
  reproducible evaluations need an exact prompt revision and model snapshot.

## Deferred

- Tag creation and the first-release contract, including behavior when `stable`
  does not yet exist.
- Exact scope-selection syntax and missing-prompt lookup behavior.
- Content equality/canonicalization rules and concurrent registration mechanics.
- Release authorization configuration, audit details, and conflict error shape.
- Prompt deletion, retention, and variable substitution beyond v1.

## Consequences

This settles Q21, Q23, and Q24 for v1. Following `stable` deliberately allows
prompt behavior to change without an application deploy; exact revisions provide
an adoption and evaluation escape hatch. Organization releases have a broad
audience, so their write permissions remain under platform control.
