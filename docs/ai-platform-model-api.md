# Model-management API

Updated: 2026-11-15
Status: Initial protobuf contract with unregistered panic-only handler stubs;
RPC behavior and authentication are not implemented.

Sources: `proto/platform/model/v1/`. Protobuf package: `platform.model.v1`.
Service: `platform.model.v1.ModelService`. Generated Go package:
`github.com/thetnaingtn/ai-system/gen/go/platform/model/v1` (package `modelv1`).

This contract covers model management, not generation. The running model process
still exposes gRPC health only. Models/tenants migrations exist; change-request
storage, mutation replay storage, permissions, and secret integration still need
implementation. See ADR-0034 for the tooling and contract decision.

## Messages and visibility

| Message | Intended audience and contents |
|---|---|
| `Model` | Safe discovery metadata: real name, provider, capabilities, lifecycle, approval, accounting mode, optional owner UUID, configuration version, timestamps |
| `ModelConfiguration` | Authorized configuration input/view, including endpoint URL; no credential or secret reference |
| `ModelCapabilities` | Four native-support booleans and positive context-token limit; fields have presence so omitted values differ from explicit false |
| `ModelChange` | Owner/platform-review view of a proposed configuration and review metadata |
| `ModelErrorDetail` | Stable error reason and safe resource/version metadata attached to a non-OK gRPC status |

Model names are globally unique real names, sent unchanged to the provider.
UUIDs are canonical UUID strings; database versions are positive signed 64-bit
integers. There are no aliases, snapshots, or caller-selected tenant IDs.

Every enum starts with `UNSPECIFIED = 0`. Unspecified accounting mode is invalid
on registration. Unspecified registration lifecycle defaults to active; registration
approval is always pending. All capability fields must be present in a complete
configuration, and the maximum context token count must be positive. These are
application validation requirements; protobuf serialization alone does not enforce
them. JSONB storage is translated to/from the typed capability message.

## RPCs

| RPC | Behavior |
|---|---|
| `RegisterModel` | Validate and reserve a name, persist a pending model/proposal, and return both. Every new registration requires platform approval, metered or unmetered. |
| `GetModel` | Return authorized discovery metadata. The default/basic view excludes endpoint configuration. An explicit configuration view requires management permission. |
| `ListModels` | Return permitted metadata in model-name order. Approval defaults to approved; pending/rejected listings require management permission and are ownership-scoped for ordinary tenants. |
| `UpdateModel` | Apply a version-checked allowed update, or return a staged proposal while leaving the active model unchanged. |
| `GetModelChange` | Return a proposal only to an authorized owner or platform reviewer. |
| `ListModelChanges` | List changes within the caller's management/review scope, ordered by creation time and ID. Optional model and approval filters narrow that scope. |
| `ApproveModelChange` | Platform-only approval, comparing the expected model revision and pending proposal state atomically; return the approved change and resulting model. |
| `RejectModelChange` | Platform-only decision with a sanitized reason. Rejecting a replacement preserves the active model; rejecting an initial registration leaves it unusable. |
| `RotateModelCredential` | Owner/platform-authorized write-only credential replacement with expected configuration version; no endpoint reapproval. |

Authorization checks run on every RPC, including mutation replays. Tenant model
ownership does not grant every workload permission to manage it. Platform-owned
registration uses an explicit flag requiring platform permission; otherwise the
owner is derived from the authenticated tenant. No request can nominate another
tenant. Missing credentials are allowed only for endpoints explicitly validated as
credentialless; an explicitly empty credential is invalid.

An identical authorized registration reuses the existing entry/proposal. A
conflicting name never overwrites the existing row. Registration does not change
ownership, reset a rejection, or expose private configuration to an unrelated tenant.
Exact secret-aware registration equality remains an implementation prerequisite.

## Updates and optimistic concurrency

`configuration.name` identifies the model. `update_mask` is required, nonempty,
and relative to `ModelConfiguration`. Writable top-level paths are `provider`,
`base_url`, `capabilities`, `status`, `sunset_at`, `replacement_model`, and
`accounting_mode`. Capability leaf paths may update individual fields, with the
complete merged configuration validated before persistence. Wildcards, unknown
paths, overlapping parent/child paths, and immutable `name` updates are rejected.

Fields outside the mask are unchanged. A masked nullable timestamp/replacement
can be cleared by omitting its value; a masked nonnullable field must have a
valid value. Setting masked capability flags to false must preserve explicit
presence. Name, owner, approval, timestamps, and version cannot be mutated by
smuggling fields into a configuration object.

`expected_config_version` is required and positive. Compare/increment in one SQL
update; never silently retry a stale mutation with a newer version. A direct
update returns the new model and no change. Endpoint URL, provider, or accounting
mode changes are staged. If a request mixes staged fields with ordinary fields,
stage the whole mutation and apply it together on approval; do not partially apply.
Use `ENDPOINT_CHANGE` for endpoint/provider changes, otherwise
`ACCOUNTING_MODE_CHANGE` for accounting-only proposals. Registration has its own type.

An optional credential on a staged endpoint/provider update belongs to that
proposal. Do not rotate the active credential before approval. The value is written
to the secret manager, never the proposal JSON, logs, or responses. Credential-only
updates use the rotation RPC. Whether pending secrets can be removed or cleared
needs an additive contract extension; empty credentials do not mean deletion.

Approvals must match the caller's expected version, the proposal's base version,
and current model version. Stale proposals need a fresh update/review. Repeated
review cannot reapply a decided proposal; safe retry uses the original request ID.
Rejecting a replacement need not match the current model version because it does
not modify the active model, but must atomically check that the proposal is pending.

The credential-rotation API stages a new managed secret version/reference and
updates the row reference/version with optimistic concurrency. Compensate or
reconcile an unused secret when the database mutation fails. An out-of-band secret
value change under an existing reference is handled by runtime credential refresh;
that external change is distinct from a version-checked rotation API call.

## Safe write retries

Each mutation requires a client-generated UUID `request_id`, scoped to the
authenticated principal, tenant, and RPC. Transport retries retain the original
ID and request. Reusing an ID with different semantic inputs returns a conflict;
never blindly reapply a mutation after a response is lost.

The handler implementation must atomically coordinate writes/replay outcomes,
including approval and rotation, with a documented retention window before these
RPCs are enabled. Secret-aware fingerprinting/replay must not persist raw secrets.
No write may inherit generation's tolerance of duplicate execution from ADR-0025.
The exact replay persistence and retention design remains open; the existing model
and tenant migrations alone are insufficient to implement these guarantees.

## Pagination

List methods default to 50 results and accept explicit sizes from 1 to 200.
Negative or excessive sizes are invalid. Tokens are opaque, tamper-resistant,
and bound to authenticated authorization scope, original filters, and ordering.
Reusing a token with changed filters/scope is invalid. Recheck authorization on
each page; pagination does not promise a cross-request database snapshot.

## Errors

Return a non-OK gRPC status, optionally with `ModelErrorDetail` for an authorized
caller. Error text/details must never echo credential values or full submitted
payloads. Do not disclose ownership/private state through unauthorized lookups.

| gRPC status | Examples |
|---|---|
| `INVALID_ARGUMENT` | Missing/invalid fields, UUIDs, enums, masks, capability values, or page tokens |
| `UNAUTHENTICATED` | Missing or invalid proven identity |
| `PERMISSION_DENIED` | Authenticated caller lacks the operation/view permission |
| `NOT_FOUND` | Model/change is absent, or resource existence is concealed from the caller |
| `ALREADY_EXISTS` | Conflicting name registration or changed inputs under a reused request ID |
| `ABORTED` | Stale model configuration version |
| `FAILED_PRECONDITION` | Change is already decided or required pricing/approval prerequisites are missing |
| `UNAVAILABLE` | Storage/secret backend temporarily unavailable; retry with the same write identity |

## Implementation boundaries

Workload identity follows ADR-0014; human portal login/delegation remains separate
design work. Generation, prompt management, policies, pricing administration,
streaming, and Go SDK convenience behavior are outside this initial contract.

Before registering handlers, implement validation, authorization, pending-change
transactions/audit, write replay, and secret handling. Endpoint checks and
permission enforcement cannot be replaced by Buf lint or schema compatibility.
