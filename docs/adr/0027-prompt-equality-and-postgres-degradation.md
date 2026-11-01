# 27. Exact prompt deduplication and cached inference during PostgreSQL outages

Date: 2026-11-01
Status: Accepted

## Decision

### Prompt equality and concurrent registration

- Compare exact registered prompt text. Do not trim whitespace, normalize newlines,
  or otherwise rewrite content before storing or comparing it.
- Calculate a content hash, using SHA-256 as the implementation default. Within the
  same scope and prompt name, identical content returns the existing immutable
  revision; any changed character, including whitespace, creates a new revision.
- Preserve the original text and compare it before reusing a hash-matched revision.
  A detected hash collision must not silently reuse another prompt's revision.
- PostgreSQL enforces unique content registration for `(scope, prompt_name,
  content_hash)` and atomically allocates distinct revision numbers for different
  content registered concurrently. Losing identical-content registrations return
  the winning revision rather than publish a duplicate.
- Registration does not move `stable` or any other release tag.
- First release requires an explicit expected-absent tag condition. If another
  publisher has created the tag, return a conflict. Later releases retain the
  expected-current-revision checks from ADR-0018.

### PostgreSQL outage behavior

- Generation may continue using already cached model catalog and tenant policy
  configuration when PostgreSQL is unavailable. Do not invent missing values or
  call discovery as a generation dependency.
- Fail requests whose required configuration is not cached, including a requested
  prompt revision that cannot otherwise be loaded. Continue to enforce available
  cached restrictions; this is not a fail-open authorization or quota policy.
- Prompt and policy writes fail until PostgreSQL recovers. Durable catalog writes
  likewise cannot succeed without their backing store.
- A cold replica with no needed configuration cannot serve those requests merely
  because other replicas have warm caches. This does not choose a readiness or
  cache-bootstrap implementation.

## Consequences and remaining work

This resolves ADR-0013's stale-cache policy and the content-equality, registration
concurrency, and first-release decisions left in ADR-0018/0023/0026. Exact text
preserves caller intent but even formatting-only edits create a new revision.
Warm replicas can continue inference through a PostgreSQL outage, while uncached
requests and writes fail clearly.

Specify the schema, revision transaction, hash-collision error, cache refresh and
coherence, maximum stale age, cold-start/readiness behavior, and visibility of
revocations during outages. Redis remains required for protected operations and
quota failure behavior is separate.
