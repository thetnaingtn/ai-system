# 8. Failover is allowed only before the first token

Date: 2026-08-30
Status: Accepted

## Context

A streaming generation can fail after the caller has already received output.
Options: fail the stream, restart on the fallback and re-emit from zero, or
disable fallback for streaming entirely.

## Decision

- **Zero bytes delivered to the caller** → transparent retry and failover are
  safe and are performed.
- **First token emitted** → the stream is **committed**. No retry, no failover.
  Terminate with `ABORTED` and a `partial: true` marker so the caller knows it
  holds a truncated response and can decide what to do.

Never re-emit from zero.

## Rationale

A caller that has already rendered 300 tokens and then receives a contradictory
restart has a corrupted result it cannot recover from at its layer. A clean
truncation with an explicit marker is recoverable; a silent contradiction is not.

## Consequences

- Streaming callers get weaker availability guarantees than unary callers. This
  is a documented property, not a bug.
- The SDK's streaming API must surface partial-response termination distinctly
  from an ordinary error.
- Deploys must drain in-flight streams (ADR-0002), since a severed stream after
  the first token cannot be recovered by failover.
