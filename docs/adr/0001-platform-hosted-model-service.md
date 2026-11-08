# 1. Model service is hosted by the platform team; the SDK is a thin gRPC client

Date: 2026-08-30
Status: Accepted

Follow-up: [ADR-0031](0031-model-names-without-snapshots.md) removes the separate
model snapshot/version-pinning concept referenced below.

## Context

Product teams across the company call LLM providers directly. Each team holds
its own provider API keys, writes its own retry and fallback logic, and there is
no central view of spend, usage, or which models are in production.

The alternative to a hosted service was an in-process library linked into each
application. That avoids a network hop and an operational burden.

## Decision

The platform team operates a **model service**. Applications talk to it through
an SDK that is a thin gRPC client: it dials the service, calls the RPC, and
returns Go values. Protobuf is the schema.

## Rationale

The primary value is **not** the uniform provider interface or the fallback
logic — those could ship as a library. It is that:

- product teams never hold provider API keys;
- spend and usage are attributable per team at the point of use;
- model routing, version pinning, and provider failover can be changed centrally
  without any application redeploying.

An in-process library gives none of these, because the decisions and the
credentials would live inside each application.

## Consequences

- The platform team is now a synchronous dependency of every AI feature in the
  company. See ADR-0002.
- A network hop and serialization cost are paid on every generation. Acceptable
  relative to LLM latency, which dominates by orders of magnitude.
- Protobuf allows non-Go clients later. The second language is not yet decided
  (see open questions); until one exists, the hop is paying for centralization
  rather than polyglot support.
- Three artifacts must be shipped and versioned: the service, the SDK, and the
  CLI (ADR-0011).
