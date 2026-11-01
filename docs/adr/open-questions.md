# Open questions

Ordered by what blocks what. Settled decisions live in the numbered ADRs.

## Proto-shaped — cheap now, breaking later

### Transport for the client edge
Deferred. Model service to internal services (data service, workflow service) is
gRPC and not in question. The client edge is open.

The strong argument for HTTP is not the usual one: having adopted OpenAI's shape
(ADR-0016), exposing `POST /v1/chat/completions` with OpenAI-compatible bodies
means the `openai` package, LangChain, LiteLLM, Instructor, notebooks and eval
harnesses point at the platform by changing one string — and Python and TypeScript
teams need no SDK from us at all.

The decision that sets the scope: **drop-in compatible** (their exact error
envelope, `finish_reason` values, SSE framing, HTTP status semantics; our own
fields — `resolution`, `degraded` — hidden in extensions their clients ignore) or
**REST-shaped with our own schema** (cleaner, and the ecosystem argument is gone).
There is no useful middle: "almost compatible" fails in a different place for every
tool.

Note grpc-gateway does not produce this. protojson base64-encodes `bytes` (breaking
`arguments`), field and enum encoding follow protobuf-JSON conventions, and
OpenAI's SSE framing including `data: [DONE]` is not something a gateway emits. An
OpenAI-compatible surface is a hand-written handler over a shared domain core.

If HTTP happens, ADR-0001 is amended: the SDK becomes the typed door for Go
services rather than the only door.

### Carried from ADR-0015 and ADR-0016
- Is structured-output emulation in v1, or does v1 ship `native | none`?
- Is reasoning effort a normalized enum or vendor passthrough?
- Is `pin` ratified as the tool-conversation fallback default?
- Behavior when a provider ignores the parallel-tool-calls toggle.

### Q15. Idempotency key
Partially settled by [ADR-0019](0019-unary-idempotency-and-recovery.md): shared
keys across unary attempts, Redis-backed result replay, 10-minute completed-result
retention, fail closed without Redis, unknown-outcome errors rather than automatic
regeneration, and continued operation after transport disconnect until its
original deadline. SDK cap clarified as two retries / three total attempts.

Further settled by [ADR-0020](0020-unary-retry-waiting-and-cancellation.md): the SDK
sends keys in gRPC metadata; changed content with the same key is rejected; retries
wait for the original running generation within the original deadline or return
its saved answer through `Generate()`. No separate retrieval API in v1. Explicit
caller cancellation triggers best-effort server/provider cancellation; repeated
cancellation is safe and completed answers remain preserved.

Still open: atomic operation ownership, cross-replica waiting/cancellation and
crash detection, fingerprint normalization, Redis record-loss behavior, provider
timeout/idempotency semantics, cancellation RPC details, and safe retry contracts
outside unary generation. Streaming replay is excluded from v1.

### Streaming API shape
Channel, iterator, or callback? How are `ABORTED` + `partial: true` (ADR-0008)
surfaced to a developer?

### Error taxonomy
Raw gRPC status codes, or typed errors (`ErrRateLimited`, `ErrContentFiltered`,
`ErrChainExhausted`)? Teams will branch on whatever ships, permanently.
ADR-0007 makes this part of the public contract.

### Scope beyond generation
Embeddings? Reranking? If they are coming, the service shape should anticipate
them.

## Correctness and safety

### Q21. Prompt x model binding
Settled for v1 by [ADR-0018](0018-prompt-registry-and-release.md): teams own
quality compatibility and evaluations. No prompt-specific model compatibility
list; per-call fallback disabling is sufficient. Technical capability enforcement
continues to apply.

### Q23. Prompt versioning
Settled by [ADR-0018](0018-prompt-registry-and-release.md): platform-assigned
immutable revisions, identical-content deduplication, default `stable` selection,
exact revision pinning, and separate releases with expected-revision conflict
checks. Tenant and platform-owned organization scopes are supported. Authorized
application backends can register and release; those permissions are separate.
First-release mechanics and deletion/retention remain open.

### Q24. Templating
Settled for v1 by [ADR-0018](0018-prompt-registry-and-release.md): store plain
system prompt text, with teams constructing changing request messages. No platform
substitution. Template support, engine choice, and variable validation are deferred
to a later version. Registered and inline system prompts are mutually exclusive.

### Logging and PII
Never asked, and the one that becomes a compliance incident. Are prompt and
response bodies logged? They contain customer data. If yes: retention, access
control, redaction. If no: how does anyone debug a bad generation?

## Operational

- **Q13. Graceful drain.** Rolling deploys sever in-flight streams unless
  `preStop` plus termination grace period exceed the longest generation.
  Kubernetes does not handle this for streams by default.
- **Q14. gRPC load balancing.** Long-lived HTTP/2 connections pin to one replica;
  naive L4 balancing produces hotspots, made worse by lazy init dialing in
  lockstep at deploy time. Client-side round-robin with a resolver, a gRPC-aware
  proxy, or a mesh?
- **Q16. Timeout alignment.** Ingress, load balancer, gRPC keepalive, and provider
  client each have idle timeouts. Any one below p99 generation time produces
  phantom mid-stream disconnects that look like provider flakiness.
- **Q25. Pricing metadata.** Who updates it when providers change prices, and does
  a missing price silently report as zero cost in chargeback?
- **Catalog write path.** Team policy goes through `policy apply` (ADR-0010). The
  platform catalog itself — reviewed config file, or an RPC? If an RPC, at
  minimum log who registered what.
- **Incident fast path.** One narrow operational verb — disable a target in a
  chain — writing straight to the store, loudly audited, auto-expiring after 24h
  so it cannot become silent config drift.

## SDK internals

### Q26. Lazy initialization
Partially settled by [ADR-0017](0017-platform-sdk-service-namespaces.md): access
to a namespace instantiates one shared gRPC client per platform instance, including
under concurrent access. Construction does not require backend readiness.

Still open: client-construction failure handling and connection recovery. A
transient failure must not accidentally become a permanently cached failure; the
precise contract has not yet been ratified.

Lazy init also trades startup failure for call-time failure: bad credentials
surface on a user-facing request instead of at boot, where the process could
crash-loop safely. What made that trade worth it?

### Q11. Second language
Never answered. Determines whether the gRPC hop is earning its keep beyond
centralization (ADR-0001), and multiplies the SDK and type-translation work.

### Proto vs domain types
Return `protoc-gen-go` types directly (the SDK is a constructor and a
pass-through), or hand-write domain types and translate (a translation layer
maintained in lockstep with every proto change, in every language)? If the
second, name the specific ergonomic gain — `oneof` handling, optional-field
semantics, hiding streaming behind an iterator.

## Not yet asked, but needed

- **Testing.** How does a team write tests without hitting real providers and
  burning tokens? A record/replay mode or a fake provider in the SDK is the
  difference between teams adopting this and teams wrapping it in their own
  abstraction.
- **Observability.** Trace propagation through gRPC metadata so an LLM call is a
  span in the caller's existing traces. Cheap by design, painful to retrofit.
- **API versioning.** How does `Generate` evolve once many teams depend on it?
- **Budget caps.** Per-team spend ceiling — deferred with the shared key
  (ADR-0003), same lever.
- **SLO and on-call.** ADR-0002 commits the platform team to being a synchronous
  dependency of every AI feature. What is the published SLO, and who carries the
  pager?
