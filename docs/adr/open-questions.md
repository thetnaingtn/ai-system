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
A retry after timeout may double-bill and, in tool-calling flows, re-execute side
effects (ADR-0009). One field in the request. Add it now.

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
A prompt tuned for one model falls back onto another, returns 200, and is quietly
worse — the worst failure mode, because nothing alerts. Does a stored prompt
declare compatible models? Does an incompatible fallback get refused, warned, or
served?

### Q23. Prompt versioning
Proposed and unratified: prompts are **immutable versions**; `latest` / `stable`
are movable tags; the response reports the resolved `prompt_version`. Rollback is
repointing a tag.

Note the change profiles are opposite: model catalog changes a few times a year
(PR is right); prompts change daily or hourly (a write API is right, or teams
will hardcode strings and the registry loses its purpose).

### Q24. Templating
Which engine, and what happens on a missing variable? Silent empty string is most
libraries' default and is wrong here — it yields a subtly broken prompt that
still returns 200. Error on missing.

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
Ten goroutines call `Generate` at boot: do all ten dial, or do nine block behind
one `sync.Once`? And does a failed dial get cached permanently by that `Once` —
turning a transient startup blip into a permanently dead client?

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
