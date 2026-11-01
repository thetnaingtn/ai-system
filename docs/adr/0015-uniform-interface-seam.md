# 15. The uniform interface seam: rename, emulate, or classify

Date: 2026-09-06
Status: Accepted

Follow-up: [ADR-0023](0023-model-feature-and-prompt-defaults.md). settles the v1 choices: defer structured-output emulation and retain provider-specific reasoning parameters. The original unratified discussion below is retained as history.

## Context

The platform promises one interface across OpenAI, Anthropic, and self-hosted
models. Every provider has features the others lack. The original framing offered
two options — lowest common denominator (the SDK is strictly worse than calling
the provider directly) or a passthrough blob (the schema guarantee is gone and
fallback becomes undefined).

Both are wrong, because the framing hides the tier that does the actual work.

## Decision

### 1. Three tiers, not two

| Tier | Example | Transfers on fallback |
|---|---|---|
| **Uniform** | messages, max_tokens, temperature, stop, tools | yes — typed fields |
| **Normalized** | system prompt, parallel-tool toggle, stop sequences | yes — we rename |
| **Passthrough** | logprobs, cache_control, service_tier, citations | no |

Most provider differences are the middle tier: the same concept spelled
differently. Those are renamed in the translation layer and never surface.

```
OpenAI    parallel_tool_calls: false
Anthropic disable_parallel_tool_use: true    -> one uniform field
```

### 2. Emulation is a separate thing from translation

Where a provider lacks a feature outright, some can still be delivered by other
means — structured output on Anthropic via a forced single tool call, on vLLM via
`guided_json`. That is not a mapping-table entry; it is code we write, maintain,
and test, with its own failure modes.

Capability support is therefore three-valued, not boolean:

```
native | emulated | none
```

The three values do not share a source, and that distinction is load-bearing
(see the Amendments section):

- **`native`** is a fact about the provider's API. It is true whether or not we
  have written any code, so it belongs in the catalog.
- **`emulated`** is a fact about *our* codebase — whether an adapter implements
  the workaround. It belongs in the adapter, not in a YAML file.
- **`none`** is neither.

`emulated` is not free — an extra tool round-trip's worth of tokens, a different
`finish_reason`, different failure modes. The response reports which mode served
the request so a team that cares can see it.

### 3. Every capability is classified by one test

> If this feature silently vanished, would the caller's code still work?

**Yes — optimization class. Drop the parameter, keep the model.**
**No — contract class. Never drop the parameter; drop the model instead.**

| Capability | Class | Behavior on an unsupporting target |
|---|---|---|
| logprobs | optimization | remove field, continue, report in `degraded[]` |
| prompt caching | optimization | remove field, continue, report |
| service tier | optimization | remove field, continue, report |
| citations | optimization | remove field, continue, report |
| structured output | contract | remove the model from the chain |
| tools | contract | remove the model from the chain |
| streaming | contract | remove the model from the chain |

Two different verbs on two different objects. One degrades the response, the
other degrades the routing. Never both for the same capability.

The classification is static, written beside the capability flag, and identical
for every tenant. It cannot be a request-time judgment: at request time we know a
field is unsupported but not what the caller does with the result. **The caller
never chooses the class** — the platform knows `logprobs` is safe to lose and
`response_format` is not, and the answer is the same for all teams.

This exists to prevent the failure named in Q21: a 200 response, green metrics,
and a caller whose `json.Unmarshal` panics because we quietly dropped the field
that guaranteed its shape.

### 4. Capability filtering happens before the first provider call

The fallback chain (ADR-0005) is filtered by the contract-class capabilities the
request needs, before any provider is contacted.

The values compared are **effective** support (§5), not raw catalog entries:

```
request needs structured_output

                 catalog   adapter emulates?   effective
gpt-4o           native    -                   native     eligible
claude-sonnet-4  none      yes                 emulated   eligible
llama-3.3-70b    none      no                  none       REMOVED
```

If the filtered chain is empty, the caller gets `FailedPrecondition` immediately,
naming the capability and the chain — before a cent is spent on a response they
cannot use.

```
FailedPrecondition: no model in the fallback chain for "gpt-4o" supports
structured_output. Chain: [llama-3.3-70b]. Fix your policy or drop the field.
```

The filter only narrows when the caller actually needs something. A request
without contract-class fields sees the full chain.

`platformctl policy validate` reports the same condition at review time — the
ADR-0006 pattern applied to capabilities rather than model names.

### 5. The catalog records native support; adapters declare emulation

**Per model, not per provider.** `gpt-4o` supports json_schema and `gpt-3.5-turbo` does not; `o1`
rejects `temperature` and system messages; Haiku and Opus differ on thinking. A
provider-level table is wrong on its first day.

The catalog holds **`native` or `none` only**. It describes the vendor's API and
makes no claim about our code.

```yaml
models:
  gpt-4o:
    snapshot: gpt-4o-2024-11-20
    provider: openai
    capabilities:
      tools: native
      structured_output: native
      streaming: true
      max_context: 128000

  claude-sonnet-4:
    provider: anthropic
    capabilities:
      tools: native
      structured_output: none      # Anthropic has no response_format field
      streaming: true
      max_context: 200000
```

Adding a model to the catalog requires filling these in — a forcing function that
comes free with ADR-0004.

Each provider adapter declares what it can deliver by other means:

```go
func (a *AnthropicAdapter) Emulates() []Capability {
    return []Capability{CapStructuredOutput}   // via forced single tool call
}

func (a *SelfHostedAdapter) Emulates() []Capability { return nil }
```

Effective support is **computed**, never stored:

```go
func effective(m CatalogEntry, c Capability) Support {
    if m.Capabilities[c] == Native {
        return Native
    }
    if slices.Contains(adapters[m.Provider].Emulates(), c) {
        return Emulated
    }
    return None
}
```

Computed once when the catalog cache loads, not per request — ADR-0006 requires
resolution to be sub-millisecond, so request time is a map read. It recomputes on
catalog refresh **and** on deploy, which is the point: a new model inherits its
provider's emulation with no code change, and new emulation reaches every model of
that provider with no catalog change.

A capability whose class is Contract (§3) and whose effective support is `None`
removes the model from the chain (§4).

### 5a. Why emulation is not a catalog value

`structured_output: emulated` in YAML is a promise that our code does something,
recorded in a file nothing verifies. Both directions fail:

- **The catalog over-promises.** The emulation was never written, or was removed in
  a refactor, and the line survives. The filter reads `emulated`, marks the model
  eligible, routes a fallback to it, and the adapter has no such code — which is
  precisely the failure §4 exists to prevent, walking past the guard.
- **The catalog under-promises.** Emulation ships, and does nothing for any model
  until a human edits every affected catalog row. Forget one and the capability is
  invisible there forever, with no signal.

Stating it in Go removes both. The declaration lives in the same file as the
implementation and is deleted by the same commit.

The general rule: **configuration describes what we do not control; code describes
what we do.**

### 6. The flag list is deliberately short

A capability flag exists **only if code branches on it**. If nothing skips a model
or emulates a feature because of a flag, it is documentation, and documentation
belongs in a description field where staleness is harmless.

v1 flags: `tools`, `structured_output`, `streaming`, `vision`, `max_context`
(needed anyway for ADR-0007's `context_length_exceeded` handling).

No flag for logprobs, caching, thinking budgets, or citations. They are
optimization class, and dropping needs no advance knowledge — the field is simply
not included for that provider.

### 7. Vendor params are keyed by provider

```protobuf
map<string, google.protobuf.Struct> vendor_params = 20;
// { "openai": {...}, "anthropic": {...} }
```

Not a flat blob. Keyed makes fallback well-defined by construction: each target
reads its own entry, a target with no entry runs on defaults, and no provider ever
receives another provider's fields.

A flat blob would send OpenAI's fields to Anthropic, Anthropic would return 400,
and ADR-0007 would read that as a provider failure — turning a schema mistake into
a routing bug.

Values are validated against a per-provider allowlist. Yes, that is a list we
maintain; the alternative is `temprature: 0.5` being silently ignored forever.
The protobuf schema guarantee is gone for this subtree, and the allowlist buys most
of it back.

### 8. Contract tests keep the matrix honest

A hand-maintained capability matrix against three vendors shipping weekly is wrong
within a month, and wrong asymmetrically:

- **Flag says `false`, reality is `true`.** We skip a model that would have worked,
  succeed on a costlier target, and **nothing alerts, ever.**
- **Flag says `true`, reality is `false`.** The provider returns 400. Loud — but
  per ADR-0007 a 400 must not trigger fallback, so the caller gets a hard failure
  carrying the provider's error text, which does not say "our catalog is wrong."

The silent one is worse and code review catches neither. A nightly job calls every
model in the catalog with every capability it claims and fails when reality
disagrees. Cheap requests, a few dollars a month. This job ships **with** the
matrix, not after it; without it the matrix degrades into folklore and we route on
folklore.

The `true`-when-false case gets its own error class so it is alertable rather than
indistinguishable from a caller's bad request.

### 9. Capability flags are public, versioned surface

`models list` (ADR-0006) exposes **effective** support, not the raw catalog value,
because teams need to know what a model will actually deliver in order to write a
sane fallback chain. Whether it arrives natively or by emulation is visible but
secondary; the per-request mode is reported in the response (§2).

The consequence is that changing effective support is a breaking change — it can
break a team's build-time lint — and it can be changed by **either** a catalog
edit or a code deploy. Shipping or removing an adapter's `Emulates()` entry is an
API change and is reviewed as one. These are versioned API, not free-form notes.

## Not yet ratified

- **Is structured-output emulation in v1?** Recommended yes: it is common, it is
  contract class, and all three provider families can do it by some means. The
  alternative is v1 shipping `native | none`, where team A's request in the flow
  above returns `FailedPrecondition` rather than being served by Claude. Cost is
  roughly a week plus a maintained code path.
- **Is reasoning effort a normalized enum** (`none|low|medium|high`, mapped to
  `reasoning_effort` or `budget_tokens` per model) or vendor-specific passthrough?
  It resembles the invented-taxonomy pattern rejected in ADR-0004, so it needs an
  explicit call rather than an assumption.
- **Prompt caching** stays passthrough for v1. Anthropic requires breakpoints at
  semantically meaningful positions and a platform guessing at placement gets it
  wrong in ways that cost money silently.

## Consequences

**The uniform surface is deliberately smaller than any single provider's.** That is
the honest position, and the escape hatch is vendor params. If the answer to every
new provider feature is "we will translate it," we are permanently behind three
vendors at once and teams route around us the moment something is missing.

**Structured output is where the SDK is better than calling the provider directly,
not merely equivalent** — one field that works across a chain of providers that
implement it three different ways. It is the counterweight to every argument that
the SDK is a tax.

**We own a translation layer that never stops moving.** Every provider feature is
a decision: rename, emulate, passthrough, or ignore. That triage is ongoing work
and should be an explicit part of the team's capacity, not a surprise.

**The caller loses a knob and that is the point.** No `importance` field, no
`allow_degradation` flag. Forty teams cannot each get the classification wrong.

## Amendments

**2026-09-13.** Original §5 recorded capability support as a single three-valued
field in the catalog, with `emulated` written in YAML. That conflated a fact about
the vendor with a fact about our codebase and allowed the catalog to promise
behavior no code implements. Amended: the catalog holds `native | none`; adapters
declare emulation in Go; effective support is computed at catalog load. §2, §4 and
§9 updated to match.
