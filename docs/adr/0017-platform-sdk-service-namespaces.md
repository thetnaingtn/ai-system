# 17. One platform SDK with lazily initialized service clients

Date: 2026-10-04
Status: Accepted (initialization failure behavior partially deferred)

Follow-up: [ADR-0021](0021-service-boundaries-and-sdk-contracts.md). resolves initialization failure policy and service boundaries; ADR-0022 chooses Envoy Gateway and automatic platform observability. The original deferred list below is retained as history.

## Context

The platform is broader than the model service. Teams within the organization
need uniform access to models, data/RAG, and sessions (messages and conversations).
The future workflow service will coordinate those primitives; applications
coordinate them until then. Tools are also exposed through a service namespace.

## Decision

- Expose service namespaces through one platform SDK, for example
  `platform.model.RegisterPrompt()` and `platform.tools.Register()`.
- Accessing a namespace lazily instantiates its gRPC client. It does not start
  the backend service or require a successful readiness check.
- Each platform instance owns one shared client per namespace. Concurrent
  accesses initialize it once and receive that same client.
- The workflow service, when introduced, owns coordination across services.

The examples describe the intended API experience, not a ratified Go signature.

## Deferred

- Exact handling of client-construction failures and connection recovery.
- Shared authentication, configuration, tracing, and error conventions across
  namespaces; namespace grouping alone does not settle these contracts.
- Detailed data, session, tools, and workflow service APIs.

## Consequences

Unused service clients need not be initialized. Lazy client construction does
not prove backend availability or valid credentials before the first RPC.
This resolves the concurrent-construction part of Q26; failure behavior remains
open. It does not settle the optional HTTP client edge.
