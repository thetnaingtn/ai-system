# 20. Unary retries wait for the original operation; cancellation is best effort

Date: 2026-11-01
Status: Accepted (distributed coordination mechanics partially deferred)

## Context

ADR-0019 established Redis-backed unary recovery and bounded operation lifetimes.
The remaining caller-facing decisions concern how retries identify and wait for
the original generation, and how explicit cancellation differs from a transport
disconnect. This record resolves those decisions and supersedes the corresponding
deferred items in ADR-0019. It does not change the three-total-attempt SDK cap or
the 10-minute completed-response retention period.

## Decision

### Identity, request validation, and response recovery

- The SDK generates one idempotency key per logical `Generate()` call and sends
  it as gRPC metadata (the transport equivalent of a header). Automatic retries
  reuse the same key and original request through the same `Generate()` API.
- Redis records are scoped by authenticated tenant and idempotency key. The same
  key in a different tenant identifies a separate operation.
- Store a request fingerprint alongside the record. If a key is reused with
  different request content, reject the request rather than return an answer
  belonging to different inputs.
- Equivalent retries of a completed operation receive its saved answer. If the
  original generation is still running, the retry waits for that generation
  instead of starting another one or returning a polling status.
- Waiting is bounded by the retry caller's remaining deadline and the original
  operation deadline. Neither retries nor waiters extend the original deadline.
- V1 has no separate operation/result retrieval API. If all SDK attempts fail,
  `Generate()` returns an error even if the service later completes and saves its
  answer. A separate new `Generate()` call normally receives a new key and starts
  a new operation.
- Automatic SDK keys protect retry attempts of one logical generation. They do
  not deduplicate separate user clicks or protect application-owned tool side
  effects. Those require action-level identity and enforcement in the API doing
  the work. Caller-supplied SDK keys have not been ratified.

### Explicit caller cancellation

- When the caller explicitly cancels its context, the SDK stops waiting and
  making generation retries, and sends a best-effort operation-cancellation
  request. This is distinct from a transport disconnect, which retains
  ADR-0019's continue-until-completion-or-original-deadline behavior.
- The cancellation request uses a separate, bounded context because the
  generation RPC's context is already cancelled. Its exact timeout remains open.
- On accepting cancellation, the service stops further retries and fallback for
  that operation and attempts to cancel the active provider request. Use the
  provider's cancellation mechanism where available, or abort the active request.
- Provider cancellation is best effort: immediate cessation of computation and
  avoidance of already incurred charges are not guaranteed.
- If the cancellation request cannot reach the service, generation can continue
  until completion or its original deadline. A local `cancel()` is not an
  acknowledgement that the remote operation has stopped.
- Repeated cancellation is safe and returns the operation's current status.
  Cancelling an operation that has already completed preserves its saved answer.

## Deferred

- Exact metadata name, fingerprint algorithm, normalization, and field coverage.
  Fingerprints must distinguish changed execution inputs without treating
  transport-only retry differences as a new request. A retry must retain the
  original operation's resolved prompt revision and model selection rather than
  re-resolve movable references and accidentally execute new work.
- Atomic operation claims, owner leases, crash detection, and the mechanism for
  another replica to wait for completion or route cancellation to the owner.
  A cancellation record alone cannot stop the owner without notification or
  observation and enforcement.
- Cancellation RPC schema, authorization details, timeout/retry policy, and
  ordering rules when completion races cancellation. Already completed results
  remain preserved as decided above.
- Retention of non-completed records, Redis record loss, provider timeout
  ambiguity, and safe retry behavior for other services remain open in ADR-0019.

## Consequences

Teams continue to call `Generate()` and use ordinary caller cancellation without
implementing a result-polling loop. The service must coordinate executing
operations independently of any one waiting RPC and across replicas. The accepted
behavior does not establish exactly-once provider execution or guaranteed remote
cancellation.
