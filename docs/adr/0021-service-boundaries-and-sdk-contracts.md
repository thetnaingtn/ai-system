# 21. Service boundaries and language-independent SDK contracts

Date: 2026-11-01
Status: Accepted

## Context

The platform is service oriented. A Go-first implementation must not restrict
which organization teams can consume its services. The user clarified service
boundaries and delegated the remaining SDK design defaults.

## Decision

### Services and consumers

- Protobuf/gRPC is the client-edge contract as well as the internal transport.
  Authorized organization teams can generate clients for their own language,
  including Go, Python, and TypeScript. V1 is not limited to Go consumers.
- An OpenAI-compatible HTTP surface is optional future work, not required for v1.
- The model service owns generation, prompts, model routing, and provider retries.
- The data service owns organization-data retrieval/RAG, sessions, conversations,
  retrieval of past conversations, and memory storage/retrieval.
- The tooling service exposes capabilities such as memory operations as tools.
  The application decides when and what to remember. Applications, or the future
  workflow service, invoke tooling; the model service does not execute tools.
- Applications compose these services initially. Workflow coordination comes
  later as its own service. These boundaries do not establish that every service
  or capability ships in the same release.
- Embedding provider integration, reranking, ingestion, and detailed data/session/
  memory APIs still need specification. RAG ownership alone does not settle these.

### SDK defaults

- Generated protobuf request/response types are the initial SDK surface. Use
  small wrappers for errors, streaming, and platform behavior rather than
  maintaining a full parallel domain type system.
- Generated RPC access is available across languages. Platform conveniences such
  as automatic keys, retry ownership, cancellation handling, and tracing require
  platform-maintained wrappers/interceptors; code generation alone does not
  implement those guarantees. The wrapper-language release matrix is still open.
- Retain one successfully initialized client per namespace and platform instance.
  Do not permanently cache a transient construction failure. Report invalid
  configuration clearly; allow transport recovery on the shared gRPC connection.
  No backend readiness call is required for namespace access.
- Expose streaming through an iterator-style wrapper with explicit close and
  terminal error access, including partial-output status. Exact method signatures
  and language-specific iterator conventions remain implementation details.
- SDK errors are typed and carry gRPC status plus structured details, allowing
  callers to branch on stable categories without parsing error text.
- Supply fake service implementations and injectable client interfaces so teams
  can test without real provider calls. Record/replay is not a v1 requirement.
- Use versioned protobuf packages, additive compatible changes within a version,
  and automated compatibility checks. Breaking contracts require a new API version.

## Consequences

This resolves the client-transport choice, Q11's language-access question, Q26's
failure policy, and SDK type, streaming, testing, and versioning defaults. V1 does
not promise a platform-maintained ergonomic SDK in every protobuf-supported
language. It promises language-independent service access.
