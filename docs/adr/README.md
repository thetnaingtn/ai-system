# Architecture Decision Records

Decisions for the central AI platform SDK and services: models, data/RAG,
sessions (messages and conversations), and tools. Model service first; the future
workflow service coordinates these primitives. Applications coordinate them
until then.

Each record states the decision, why, and what it costs. A decision that turns
out wrong gets a new ADR that supersedes the old one — records are never edited
to hide history.

| # | Decision | Status |
|---|---|---|
| [0001](0001-platform-hosted-model-service.md) | Model service is hosted by the platform team; the SDK is a thin gRPC client | Accepted |
| [0002](0002-hard-fail-not-client-bypass.md) | Service unavailability is a hard fail; clients never bypass to providers | Accepted |
| [0003](0003-shared-provider-key.md) | One shared provider key per provider (noisy neighbor accepted) | Accepted |
| [0004](0004-flat-catalog-family-to-snapshot.md) | Flat org-wide model namespace; families map to pinned snapshots | Accepted |
| [0005](0005-per-team-fallback-policy.md) | Fallback chains are per-team config, not request payload | Accepted |
| [0006](0006-discovery-vs-resolution.md) | Discovery is build-time; resolution is request-time; separate code paths | Accepted |
| [0007](0007-fallback-triggers.md) | Which failures trigger fallback, and which never do | Accepted |
| [0008](0008-streaming-failover-boundary.md) | Failover is allowed only before the first token | Accepted |
| [0009](0009-retry-policy.md) | Retry defaults, team-tunable within platform caps, deadline-bounded | Accepted |
| [0010](0010-config-authoring-and-distribution.md) | Teams author policy in their own repos; CI pushes to the store | Accepted |
| [0011](0011-platformctl-cli.md) | A thin CLI is the build-time interface for humans and CI | Accepted |
| [0012](0012-request-knobs-restrict-only.md) | Request-level knobs may restrict behavior, never expand it | Accepted |
| [0013](0013-storage-and-degradation.md) | Shared store with in-process read-through cache | Partially deferred |
| [0014](0014-tenant-identity.md) | Tenant identity is proven by the platform, resolved at the transport boundary | Accepted |
| [0015](0015-uniform-interface-seam.md) | The uniform interface seam: rename, emulate, or classify | Accepted |
| [0016](0016-tool-calling.md) | Tool calling: OpenAI's shape, pinned and owned | Accepted |
| [0017](0017-platform-sdk-service-namespaces.md) | One platform SDK with lazily initialized service clients | Accepted; failure behavior partially deferred |
| [0018](0018-prompt-registry-and-release.md) | Immutable prompt revisions with separate, authorized releases | Accepted |
| [0019](0019-unary-idempotency-and-recovery.md) | Unary retry recovery uses Redis records and bounded operations | Accepted; cancellation and mechanics partially deferred |
| [0020](0020-unary-retry-waiting-and-cancellation.md) | Unary retries wait for the original operation; cancellation is best effort | Accepted; coordination mechanics partially deferred |

See [open-questions.md](open-questions.md) for what is still undecided.
