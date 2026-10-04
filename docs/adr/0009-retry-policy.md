# 9. Retry defaults, team-tunable within platform caps, deadline-bounded

Date: 2026-08-30
Status: Accepted

Later clarification and partial supersession:
[ADR-0019](0019-unary-idempotency-and-recovery.md) defines the SDK cap as two
retries (three total attempts), requires protection for uncertain execution, and
supersedes cancellation propagation for unary transport disconnects. The original
decision below is retained as history; explicit caller cancellation remains open.

## Context

Retry is nested inside fallback. Done naively the two multiply:

    SDK retries 3 x service retries 3 per target x 3 targets = 27 provider calls

During a provider brownout — exactly when every client retries at once — that
turns a degradation into an outage, on a shared key, for the whole org.

## Decision

### Where retry lives

- **The service owns provider retries.** It has the error taxonomy, the circuit
  breakers, and the cross-tenant budget.
- **The SDK retries only `UNAVAILABLE` from the model service** (connection-level,
  request never landed), capped at 2, using gRPC's built-in retry policy.
- Never both.

### Defaults

```yaml
retry:
  max_attempts_per_target: 2        # 1 try + 1 retry
  backoff:
    initial: 200ms
    multiplier: 2
    max: 5s
    jitter: full
  respect_retry_after: true
  max_retry_after_wait: 10s
  per_attempt_timeout: 30s
  total_deadline: 60s
```

Two attempts, not three: with a fallback chain behind it, a second target is a
better use of the next interval than a third try at a sick provider.

Full jitter is not optional. Without it, clients that failed at the same instant
retry at the same instant, in waves, and nothing recovers.

### Team-tunable, with hard caps

| Field | Team may set | Cap |
|---|---|---|
| `max_attempts_per_target` | 1–3 | 3 |
| `per_attempt_timeout` | 5s–60s | 60s |
| `total_deadline` | 5s–120s | 120s |
| backoff / jitter | not configurable | — |

Caps are enforced server-side at `policy apply`, not only in CI. Because the
provider key is shared (ADR-0003), one team's retry configuration is everyone's
problem.

### The deadline is the budget

The loop checks remaining time **before every attempt** and clamps each attempt
to the time actually left:

```go
remaining := time.Until(deadline)
if remaining < minUsefulTime {
    return errDeadline
}
attemptCtx, cancel := context.WithTimeout(ctx, min(perAttemptTimeout, remaining))
```

Without this, `3 attempts x 30s x 3 targets` = a caller waiting 270s on a 60s
deadline. Everything past the 60s line is work done for a caller who has already
given up: tokens generated and paid for that nobody reads, and a goroutine,
connection, and provider concurrency slot held 4.5x longer than they should be —
during an outage, which is when the service can least afford it. The code that
tries hardest to be resilient is what takes the service down.

### Context propagation

The caller's `context.Context` must reach the provider HTTP call unbroken. A
`context.Background()` anywhere in the chain — typically in a helper that has no
`ctx` parameter — silently severs cancellation. The in-flight attempt then runs
to its own timeout, or forever, and the provider keeps generating billable
tokens for a caller who has gone.

Enforcement:

- lint `context.Background()` out of everything but `main()` and tests
  (`forbidigo` / `contextcheck`);
- a test that cancels the caller at 100ms and asserts the **fake provider
  observed the cancellation**, end to end.

### Protecting the provider

- **Circuit breaker per (provider, model).** Open after N consecutive failures,
  skip the target during cooldown, half-open probe to test recovery. Without it,
  every request pays a full timeout against a dead provider before falling
  through, and p99 goes to the deadline even though a healthy fallback exists.
- **Retry budget**, not just a count: cap retries at ~10% of request volume to a
  target. Per-request limits bound one request's blast radius; only the budget
  bounds aggregate retry load on a provider that is already drowning.

## Consequences

- Retry configuration is part of team policy and is validated on apply.
- Idempotency keys are required for timeout-retry to be safe (open question).
- Streaming retry is bounded by ADR-0008: before the first token only.
