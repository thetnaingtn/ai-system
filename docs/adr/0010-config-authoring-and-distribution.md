# 10. Teams author policy in their own repos; CI pushes to the store

Date: 2026-08-30
Status: Accepted

## Context

Per-team fallback policy (ADR-0005) has to reach the model service at request
time. That does not mean the files live in the model service repo.

## Decision

Three separate layers:

```
AUTHORING                  DISTRIBUTION              RUNTIME
team-a/config/llm.yaml ─┐
   (team A's repo)      │  team CI on merge:        model service
                        ├─► platformctl apply ────► in-process cache
team-b/config/llm.yaml ─┘        │                       ▲
   (team B's repo)               ▼                       │
                          Postgres/etcd ─────────────────┘
                            (versioned)
```

- **Authoring** — the team's own repo, next to their code, under their
  CODEOWNERS, reviewed by them.
- **Distribution** — **push**: team CI runs `platformctl policy apply` on merge.
- **Runtime** — the service reads from cache backed by the store. It never reads
  a file from anyone's repo.

Push rather than pull: pull requires the platform to maintain a registry of repo
URLs that goes stale on every rename, and takes control of timing away from the
team.

## Validation happens twice

CI validation is fast feedback and is bypassable (skipped pipeline,
misconfigured job, a hand-run `apply`). The **server re-validates on apply** and
rejects: unknown or sunset model names, providers outside the tenant allowlist,
capability-incoherent chains, retry values above caps, and any attempt to write
another tenant's policy.

Same policy engine, run in both places. CI is UX; the server is enforcement.

## Consequences

- Policies are versioned on write, so `policy rollback` is a store operation —
  instant, and it does not require the team's CI to be healthy, which matters at
  3am when their pipeline is as broken as everything else.
- A new team with no published policy gets **no fallback**: calls work, and fail
  cleanly rather than silently inheriting invented behavior.
- Pushing requires the team to authenticate — the same tenant identity that
  authorizes `Generate` (open question).
- **Locality is preserved.** The chain sits in the team's repo next to their
  code, which was the main appeal of putting it in the request body.

## Pragmatic v1

With a handful of teams, or in a monorepo, put every team's policy in one file
in the platform repo and merge their PRs. The runtime contract is `store →
service` either way, so federating authoring later changes only who opens the PR
and which CI pushes — no client change, no proto change, no migration.

Federate when PR volume becomes annoying. That is the signal, not a headcount.
