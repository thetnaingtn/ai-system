# 19. Unary retry recovery uses Redis records and bounded operations

Date: 2026-10-04
Status: Accepted (explicit cancellation and recovery mechanics partially deferred)

Follow-up: [ADR-0020](0020-unary-retry-waiting-and-cancellation.md) resolves the
explicit cancellation contract, metadata-based key transport, rejection of changed
request content, waiting for running operations, and recovery through `Generate()`
without a separate retrieval API. The original deferred items below are retained
as history; distributed coordination mechanics remain open.

## Context

A service can finish a generation but lose its connection before delivering the
response. Retrying without operation identity may perform another billable
generation. Avoiding duplicate execution alone is insufficient: the retry also
needs a way to receive the original result.

## Decision

### Request attempts

- Clarify ADR-0009's SDK cap: two retries means **three total attempts**, one
  initial call plus retry 1 and retry 2, bounded by the caller's deadline.
- The existing SDK retryable status remains eligible `UNAVAILABLE` failures.
  That status alone is not evidence that the request never executed.
- The provider-side default remains two total attempts per target, with the
  existing configurable cap of three. SDK and provider attempt limits are
  distinct.
- A logical retried unary operation carries the same idempotency key across its
  attempts. A separate new operation has a separate key. The SDK supplies keys
  for its automatic retry path.

### Redis records and result delivery

- Redis holds temporary operation records and completed unary responses, shared
  across service replicas and scoped by authenticated tenant and idempotency key.
  It is not the durable prompt registry or catalog store.
- Save a completed response before delivering it. A retry with the same key
  retrieves that response and receives it through the retry RPC, without needing
  the original connection.
- A retry of an operation still running joins the existing operation rather than
  starting another generation.
- Completed responses remain available for **10 minutes after completion**.
- Requests relying on idempotency protection fail before invoking the provider
  when Redis is unavailable; do not silently execute without protection.
- Retention must protect records from automatic eviction during their lifetime.
  Redis loss or record expiry ends the available deduplication guarantee; using
  Redis does not establish exactly-once provider execution.
- If a replica crashes after provider completion but before saving the result,
  and the outcome cannot be recovered, return an explicit **outcome unknown**
  error. Do not automatically regenerate; the caller decides whether to start a
  new operation.

### Operation lifetime and streaming

- For unary generation, a transport disconnect does not cancel the underlying
  operation. Continue until completion or the original operation deadline.
- A retry does not reset or extend that deadline.
- This supersedes ADR-0009's unconditional propagation of RPC cancellation for
  this unary transport-disconnect case. Operations still require a bounded
  lifetime; it is not permission for unbounded background generation.
- V1 completed-result replay is unary only. Streaming retains ADR-0008's
  commitment boundary: no automatic retry or failover after output begins.

## Deferred

- Explicit caller cancellation: whether the SDK translates `cancel()` into a
  best-effort cancellation RPC, the cancellation API, and its acknowledgement
  semantics. This was discussed but not accepted.
- Atomic claim/completion rules, request fingerprint validation, owner leases,
  and how other replicas join running work or detect unrecoverable outcomes.
- Redis persistence/failover configuration, record-loss behavior, and retention
  of running, failed, cancelled, and unknown-outcome records.
- Handling retries that arrive after all SDK attempts are exhausted or records
  expire; a separate result-retrieval API has not been accepted.
- Provider idempotency support and provider timeout/failover ambiguity. A Redis
  record does not automatically prevent duplicate provider charges.
- Safe retry contracts for other services and write methods.

## Consequences

This settles the principal unary recovery choices in Q15, but implementation and
cancellation contracts remain open. Redis becomes a required dependency for
protected operations. Disconnect recovery can consume tokens after a caller has
left, bounded by the original deadline. An idempotency key does not protect
side effects performed by application-owned tools.
