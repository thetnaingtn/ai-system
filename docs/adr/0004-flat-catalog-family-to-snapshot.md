# 4. Flat org-wide model namespace; families map to pinned snapshots

Date: 2026-08-30
Status: Accepted

Follow-up: [ADR-0031](0031-model-names-without-snapshots.md) supersedes the snapshot
requirements below. The flat namespace and catalog lifecycle remain accepted;
historical snapshot decisions are preserved.

## Context

Callers need to name a model. Two questions: what namespace do names live in,
and how specific is the name?

An earlier draft used invented intent names (`summarize-cheap`,
`summarize-fast`). That was rejected: intent taxonomies require org-wide
vocabulary agreement, go stale, and force developers to translate what they
actually want ("gpt-4o") into someone else's noun.

## Decision

- **One flat, org-wide namespace.** Names are globally unique. There is no
  per-team prefix. The catalog is curated and shared: any team may use any
  entry.
- **Callers name real models by family** (`gpt-4o`, `claude-sonnet`,
  `llama-70b`). The catalog maps each family to a currently pinned vendor
  snapshot (`gpt-4o-2024-11-20`).
- **Dated snapshots do not appear in application code**, except as a deliberate,
  visible escape hatch for teams doing strict eval reproducibility.
- Catalog entries carry `status` (ACTIVE / DEPRECATED / SUNSET), `sunset_at`,
  `replacement`, capabilities, and pricing metadata.

## Rationale

Vendor model ids necessarily end up hardcoded in application source. There are
two kinds, with very different churn:

| | example | churn |
|---|---|---|
| family | `gpt-4o` | rare — lives for years |
| snapshot | `gpt-4o-2024-11-20` | frequent — vendors rotate on a schedule |

Confining hardcoded names to families converts a recurring org-wide migration
(every snapshot rotation) into a rare one (family retirement).

## Consequences

- **Snapshot rotation costs zero diffs.** One catalog edit, canary at a small
  percentage, roll forward or revert in seconds. No team redeploys.
- **No silent drift.** Without the mapping, "we use gpt-4o" is quietly several
  different snapshots across the org with different behavior and price.
- **The response must report the real snapshot**, not the family, so teams can
  correlate an output-quality change with a rotation. See ADR-0006.
- Family retirement still requires teams to ship a change. The cost of *finding*
  them is removed by deprecation metadata plus CI lint (ADR-0006), not by this
  ADR.
- A globally unique flat namespace means first-come-first-served on names.
  Accepted deliberately; the catalog is small and curated.
