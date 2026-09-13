# 5. Fallback chains are per-team config, not request payload

Date: 2026-08-30
Status: Accepted

## Context

Teams have different priorities. One team wants OpenAI primary with Anthropic
behind it; another wants the reverse; a third cannot use a self-hosted model at
all for data-residency reasons. Teams should control their own fallback.

The question was *where* that choice is expressed: in the request body, or in
registered configuration.

## Decision

Fallback chains live in **per-team configuration**, keyed on the real model name
the caller asked for. The request carries a model name; the service resolves the
chain from the calling team's policy.

```yaml
# team-a
fallbacks:
  gpt-4o: [claude-sonnet, llama-70b]

# team-b
allowed_providers: [openai, anthropic]
fallbacks:
  gpt-4o:        [claude-sonnet]
  claude-sonnet: [gpt-4o]
```

Team A and Team B send an identical request and get their own behavior.

## Rationale

The requirement is **per-team** variation, not **per-request** variation. It was
confirmed that a team's preferred chain does not need to change between two
requests from that team. Per-team config satisfies the requirement completely.

Chains in the request body would additionally mean:

- **Speed.** Config changes in about a minute. Source changes are a deploy,
  across N teams on N release schedules — weeks, on a timeline the platform does
  not control. Providers do not fail on anyone's release schedule.
- **Consent.** Once teams hold the value, routing cannot change without their
  cooperation. Every change becomes a migration campaign.
- **Inventory.** "Who routes to Anthropic?" becomes unanswerable without grepping
  every repo, and copy-pasted chains drift.

## Consequences

- Teams own their own policy and change it without platform approval
  (ADR-0010). Self-service is a **requirement**: if registering a policy is slow,
  teams will avoid the feature and the platform loses its leverage.
- A missing policy entry means **no fallback** — fail cleanly. The platform never
  invents a chain on a team's behalf.
- Provider allowlists are enforced server-side, both at `policy apply` and at
  request time.
- Capability coherence of a chain is validated at config time, not discovered
  during a failover.
- Dynamic per-request choice remains available by naming a different model.
