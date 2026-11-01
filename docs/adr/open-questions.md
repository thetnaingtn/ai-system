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

## Remaining design and implementation work

- **Durable storage and degradation (ADR-0013):** PostgreSQL is selected for
  durable records and Redis for temporary state (ADR-0026). Specify schema and
  transactions, cache refresh/coherence, maximum stale age, cold-start/readiness,
  and revocation visibility. ADR-0027 settles serving cached configuration while
  PostgreSQL is down, failing on missing configuration, and rejecting writes.
- **Unary recovery (Q15):** define atomic claim/completion transitions, owner lease
  durations, fingerprint normalization, waiter notification, cancellation routing,
  and completion/cancellation race ordering. Specify takeover behavior,
  non-completed-record retention, Redis failover/record loss, and provider
  timeout/idempotency behavior. ADR-0025 permits bounded regeneration after lost
  records or expired owners, accepting duplicate provider cost and changed output.
  No PostgreSQL execution ledger is required solely to prevent duplicates.
  Determine safe retry contracts for service writes beyond generation separately.
- **Tenant protection:** choose numeric quotas, token/concurrency accounting,
  reservation implementation and reconciliation (including uncertain generation
  costs), production capacity allocation, month-boundary attribution, and outages
  of enforcement state. ADR-0028 settles monthly UTC resets, rejection on
  exhaustion with in-flight completion, missing-price blocking, and approved
  unmetered exemptions. Do not silently revert to unlimited use.
- **Data, sessions, memory, and tools:** specify schemas, access boundaries,
  lifecycle/retention, tool invocation, and service-owned storage contracts.
  Determine embedding integration, ingestion, and reranking scope and release
  sequencing after the first model service (ADR-0026). The application owns
  deciding when to write memory.
- **SDK experience:** Go is the initial platform SDK language (ADR-0026); other
  generated gRPC clients remain supported at the service boundary. Finalize Go
  packaging, streaming signatures, and typed errors. Additional language wrappers
  are deferred; generated clients alone do not implement automatic platform behavior.
- **Prompt details:** implement hash-based exact-text equality, PostgreSQL
  uniqueness/revision transactions, collision handling, expected-absent first-release
  encoding, reference tracking, and later physical deletion design. ADR-0027 settles
  the equality, concurrent registration, and first-release behavior.
- **Custom models and secrets:** implement organization-wide approved entries,
  pending unmetered approval, staged endpoint changes, owner/platform update
  authorization, and managed secret references (ADR-0029). Finalize registration
  equality, endpoint validation, secret backend and rotation refresh, approval
  auditing, and reconciliation with CI-managed baseline catalog entries.
- **Deployment and telemetry:** Envoy route/security policies, timeouts and drain
  margins, trace sampling/exporters, safe metadata allowlists, telemetry retention,
  and monitoring for best-effort operations. Ensure logging/exception paths cannot
  emit content prohibited by ADR-0022.

Some of these details may reveal product tradeoffs. The delegated defaults do not
silently authorize changes to the accepted scope, privacy, or availability policy.
