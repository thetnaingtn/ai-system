# Open questions

Updated: 2026-11-01

The five product/operating questions discussed with the owner have been answered.
The remaining technical defaults were delegated and are recorded in ADR-0021
through ADR-0024. This does not mean every implementation contract or service API
is complete. The items below distinguish settled policy from remaining design work.
Follow-up decisions in [ADR-0025](0025-performance-first-generation-recovery.md)
accept occasional duplicate generation during recovery; [ADR-0026](0026-storage-and-model-first-delivery.md)
selects PostgreSQL/Redis, Go SDK delivery, and model-service-first sequencing.
[ADR-0027](0027-prompt-equality-and-postgres-degradation.md) settles prompt equality
and PostgreSQL outage behavior; [ADR-0028](0028-monthly-budgets-and-unmetered-model-approval.md)
settles monthly budgets and approved unmetered models; [ADR-0029](0029-custom-model-registration-and-secrets.md)
settles organization-wide custom models, endpoint updates, and managed credentials.

## Settled policy and technical defaults

| Original topic | Decision / record |
|---|---|
| Client transport | Protobuf/gRPC across languages; optional HTTP compatibility later. [ADR-0021](0021-service-boundaries-and-sdk-contracts.md) |
| Structured-output emulation | Deferred beyond v1; native support only initially. [ADR-0023](0023-model-feature-and-prompt-defaults.md) |
| Reasoning effort | Provider-specific passthrough. [ADR-0023](0023-model-feature-and-prompt-defaults.md) |
| Tool-conversation fallback | Pin by default; explicit tenant-policy opt-out. [ADR-0023](0023-model-feature-and-prompt-defaults.md) |
| Parallel-tool restriction | Typed error on violation before exposing executable calls. [ADR-0023](0023-model-feature-and-prompt-defaults.md) |
| Q15: idempotency | Same-key retry recovery, Redis result retention, waiting, explicit best-effort cancellation, and no retrieval API. [ADR-0019](0019-unary-idempotency-and-recovery.md), [ADR-0020](0020-unary-retry-waiting-and-cancellation.md); coordination defaults in [ADR-0024](0024-v1-operating-policy-and-recovery-defaults.md) |
| Streaming API | Iterator-style wrapper with close, terminal errors, and partial-output reporting. [ADR-0021](0021-service-boundaries-and-sdk-contracts.md) |
| Error taxonomy | Typed SDK errors plus gRPC status and structured details. [ADR-0021](0021-service-boundaries-and-sdk-contracts.md) |
| Scope beyond generation | Model service remains focused; RAG, sessions, conversations, and memory belong to data, with tooling exposure and application-owned memory decisions. [ADR-0021](0021-service-boundaries-and-sdk-contracts.md) |
| Q21: prompt/model quality compatibility | Team-owned; per-call fallback disabling is sufficient for v1. [ADR-0018](0018-prompt-registry-and-release.md) |
| Q23: prompt versioning | Immutable revisions, separate conflict-checked releases, default stable tag. [ADR-0018](0018-prompt-registry-and-release.md); explicit first release and scope lookup in [ADR-0023](0023-model-feature-and-prompt-defaults.md) |
| Q24: templating | No platform substitution in v1. [ADR-0018](0018-prompt-registry-and-release.md) |
| Logging and PII | Prohibit content in diagnostic logs/traces; keep functional storage separate. [ADR-0022](0022-envoy-gateway-observability-and-content-privacy.md) |
| Q13: graceful drain | Drain streams and disconnected running operations within deadlines. [ADR-0024](0024-v1-operating-policy-and-recovery-defaults.md) |
| Q14: load balancing | Envoy Gateway owns backend routing/balancing. [ADR-0022](0022-envoy-gateway-observability-and-content-privacy.md) |
| Q16: timeout alignment | Original operation deadline, bounded attempts, infrastructure margin. [ADR-0024](0024-v1-operating-policy-and-recovery-defaults.md) |
| Q25: pricing metadata | Platform-owned reviewed updates; missing price is unknown, not zero. [ADR-0024](0024-v1-operating-policy-and-recovery-defaults.md) |
| Catalog write path | Reviewed baseline configuration via CI, plus authenticated custom-model registration/update APIs. [ADR-0024](0024-v1-operating-policy-and-recovery-defaults.md), [ADR-0029](0029-custom-model-registration-and-secrets.md) |
| Incident fast path | Authorized, audited target-disable override expiring after 24 hours. [ADR-0024](0024-v1-operating-policy-and-recovery-defaults.md) |
| Q26: lazy initialization | One shared client per namespace; recover transient failures; clear configuration errors. [ADR-0017](0017-platform-sdk-service-namespaces.md), [ADR-0021](0021-service-boundaries-and-sdk-contracts.md) |
| Q11: second language | Any supported protobuf/gRPC client language can call services; wrapper coverage is separate. [ADR-0021](0021-service-boundaries-and-sdk-contracts.md) |
| Proto versus domain types | Generated types plus small behavior/stream/error wrappers. [ADR-0021](0021-service-boundaries-and-sdk-contracts.md) |
| Testing | Fake services and injectable client interfaces. [ADR-0021](0021-service-boundaries-and-sdk-contracts.md) |
| Observability | Automatic platform instrumentation and standard trace propagation, with optional application custom spans. [ADR-0022](0022-envoy-gateway-observability-and-content-privacy.md) |
| API versioning | Versioned protobuf packages, additive evolution, compatibility checks. [ADR-0021](0021-service-boundaries-and-sdk-contracts.md) |
| Budget caps | Monthly UTC spend caps with reservations, rejection of new work on exhaustion, and reviewed unmetered exemptions; separate rate/concurrency protection. [ADR-0028](0028-monthly-budgets-and-unmetered-model-approval.md) |
| SLO and on-call | Best-effort v1; no contractual availability promise or 24/7 commitment. [ADR-0024](0024-v1-operating-policy-and-recovery-defaults.md) |

## First model service: remaining design and implementation work

Major architecture and policies are settled. Most items below are implementation
choices delegated to the implementer, not blockers requiring another owner
interview. Record the selected details as implementation proceeds; raise new
product tradeoffs if they change accepted scope, privacy, cost, or availability.

Start with the protobuf/Go contract and PostgreSQL schema, then implement the
model service against the accepted ADRs.

### 1. API contracts and Go SDK

- [ ] Finalize versioned protobuf messages and RPCs for generation/streaming,
  prompts/releases, discovery, policies, model registration/approval/update,
  credential rotation, and explicit cancellation.
- [ ] Define structured error details and typed Go SDK errors.
- [ ] Finalize Go package layout, generated types, lazy client construction,
  iterator/close behavior, retry metadata, cancellation, and test interfaces.
- [ ] Preserve cross-language generated gRPC access. Go is the initial maintained
  platform SDK; additional language wrappers are deferred (ADR-0026).

### 2. PostgreSQL design and prompt mechanics

- [ ] Define tables, migrations, transactions, indexes, ownership/authorization
  records, and constraints for catalog, policies, prompts/releases, approvals,
  pricing, and budget-related durable records.
- [ ] Implement SHA-256-based exact-text prompt deduplication with stored-text
  comparison, collision handling, and concurrent uniqueness enforcement.
- [ ] Allocate distinct revisions atomically for changed content; return existing
  revisions for identical content without moving release tags.
- [ ] Encode expected-absent first releases and expected-current-revision updates,
  returning conflicts on races (ADR-0027).
- [ ] Implement prompt reference tracking. Physical deletion remains deferred.

### 3. Cache behavior and PostgreSQL outages

- [ ] Specify refresh timing, configuration coherence, maximum stale age, and
  visibility of authorization/policy revocations during outages.
- [ ] Define cold-start/bootstrap and readiness behavior.
- [ ] Preserve the settled policy: serve cached required configuration, fail on
  missing required configuration, and reject writes during PostgreSQL outages.
  Redis and quota availability rules remain separate (ADR-0027).

### 4. Monthly budgets and tenant protection

- [ ] Choose numeric rate, token, concurrency, and spend limits and production
  capacity allocation.
- [ ] Implement cost reservation, actual-usage reconciliation, retry/recovery
  accounting, and handling of uncertain or missing provider usage/cost.
- [ ] Specify month-boundary attribution, reservation expiry/release, and behavior
  when enforcement state is unavailable.
- [ ] Preserve monthly UTC reset, rejection of new work on exhaustion, in-flight
  completion, and paid-model missing-price blocking. Only approved unmetered
  models bypass provider-spend accounting; rate/concurrency limits still apply
  (ADR-0028).

### 5. Redis recovery and cancellation

- [ ] Define atomic claim/completion transitions, request fingerprint fields and
  normalization, ownership tokens, lease durations, and takeover timing.
- [ ] Implement cross-replica waiting/completion notification and cancellation
  delivery/observation; settle completion-versus-cancellation ordering.
- [ ] Define retention of non-completed records, Redis failover/record-loss
  behavior, and original-deadline/attempt-budget propagation across takeovers.
- [ ] Preserve best-effort generation deduplication: bounded regeneration after
  lost records or expired owners is allowed, even if it duplicates provider cost
  or changes the answer. No PostgreSQL execution ledger is required solely to
  prevent duplicate generation (ADR-0025).
- [ ] Define safe retry contracts for service writes independently; generation
  tolerance does not authorize replay of application-owned tool side effects.

### 6. Custom models and secret management

- [ ] Finalize registration equality, approval schema, owner/platform update
  authorization, capability validation, and endpoint destination validation.
- [ ] Implement organization-wide approved models and pending, unusable no-price
  registrations until unmetered approval.
- [ ] Stage replacement endpoints for reapproval while keeping the current
  approved endpoint active; rejection preserves the active endpoint.
- [ ] Select the secret-manager backend (Vault is a candidate), permissions,
  write-only credential input, references, rotation refresh, and partial-write
  cleanup. Rotation requires no consumer application redeploy.
- [ ] Specify approval auditing/concurrency and reconciliation of runtime custom
  entries with CI-managed baseline catalog configuration (ADR-0029).

### 7. Deployment and observability

- [ ] Configure Envoy Gateway routes, backend discovery/balancing, authentication,
  forwarded metadata, and timeout settings; disable gateway generation retries.
- [ ] Define shutdown/readiness timing and drain active streams plus disconnected
  unary operations within original deadlines.
- [ ] Configure automatic trace propagation/instrumentation, sampling, exporters,
  safe metadata allowlists, telemetry retention, and best-effort monitoring.
- [ ] Verify logs, traces, gateway access logs, and provider errors cannot emit
  prompt/document/conversation/response bodies or credentials (ADR-0022).

## Deferred beyond the first model service

- Data/RAG, sessions, historical conversations, memory, and tooling APIs: schemas,
  access boundaries, retention, storage, tool invocation, embedding integration,
  ingestion, and reranking scope. The application decides when to record memory;
  detailed discussion follows the first model service (ADR-0026).
- Workflow-service orchestration.
- Additional maintained language SDK wrappers, optional OpenAI-compatible HTTP
  access, prompt substitution, structured-output emulation, and physical prompt
  deletion, as deferred by the accepted ADRs.
