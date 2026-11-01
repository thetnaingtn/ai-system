# 16. Tool calling: OpenAI's shape, pinned and owned

Date: 2026-09-06
Status: Accepted

Follow-up: [ADR-0023](0023-model-feature-and-prompt-defaults.md). ratifies pinning as the default and requires a typed error when a provider violates a requested parallel-tool restriction. The original unratified items below are retained as history.

## Context

Tool calling is the hardest part of a uniform interface. Tool schemas, tool-call
ids, argument encoding, and multi-turn state all differ across providers. Whatever
ships becomes permanent: teams write tool loops once and keep them for years.

## Decision

### 1. OpenAI's shape is the canonical shape

Callers write OpenAI-shaped tools. The platform translates to Anthropic and to
anything else.

The deciding argument is not familiarity. **vLLM, Ollama, TGI, Together, and Groq
already serve OpenAI-compatible APIs**, so OpenAI's shape means one real
translation layer (Anthropic) and pass-through for most of the catalog. Anthropic's
shape would mean translating for the majority.

| | OpenAI | Anthropic |
|---|---|---|
| declaration | `function.parameters` | `input_schema` |
| choice | `tool_choice: auto\|none\|required\|{name}` | `tool_choice: {type: auto\|any\|tool}` |
| model emits | `tool_calls[].function.arguments` (JSON **string**) | `tool_use` block, `input` (**object**) |
| caller returns | `role: "tool"`, `tool_call_id` | `tool_result` block, `tool_use_id` |

All of it is mechanical. None of it is the caller's problem.

### 2. Pinned and owned, not tracked

We adopt OpenAI's shape **as of a pinned date, as our own schema.** It looks like
OpenAI's because that is the most familiar and most widely compatible shape, not
because we follow their releases.

OpenAI deprecated `function_call` in favour of `tool_calls` and will do something
like it again. Tracking their spec means either breaking our teams on OpenAI's
schedule, or quietly ceasing to be "the OpenAI standard" while still claiming to
be. Owning the shape means a future OpenAI divergence is handled in the OpenAI
provider adapter, like any other provider. The `.proto` does not move.

Same wire format on day one; a completely different position in year two.

### 3. Arguments are validated JSON, not a possibly-malformed string

OpenAI returns arguments as a JSON string that can be truncated on `max_tokens` or
simply invalid from a weaker model. Anthropic returns a parsed object. Passing
OpenAI's string through would export that failure to every team, including those
calling models where it cannot happen — and the same tool call would arrive in two
shapes depending on which target served it.

```protobuf
message ToolCall {
  string id        = 1;
  string name      = 2;
  bytes  arguments = 3;  // JSON. Guaranteed to parse and match the declared schema.
}
```

`bytes` rather than `Struct` deliberately: `json.Unmarshal(tc.Arguments, &args)` is
exactly what a caller writes against OpenAI today, so the migration diff is zero.
`structpb.Struct` would force `AsMap()` and a second conversion.

The platform parses once and validates against the declared schema. Malformed
output becomes a typed error at our boundary — `ErrToolArgumentsInvalid`, raw text
attached — instead of forty teams each writing repair logic.

**This is the one deliberate deviation from OpenAI's shape.** We take the
structure and reject the stringly-typing. Over an HTTP surface the same value is
emitted as an OpenAI-compatible string; the guarantee is that it always parses.

### 4. `strict` is a vendor param, not part of the canonical schema

OpenAI's `strict: true` guarantees schema adherence but only over a subset of JSON
Schema: every property required, `additionalProperties: false`, limited keyword
support. Putting it in the canonical schema imposes that subset on every team,
including those targeting models where it buys nothing.

Canonical `parameters` is plain JSON Schema. `strict` lives in vendor params
(ADR-0015 §7), and using it makes a tool schema provider-specific.
`platformctl policy validate` warns when `strict: true` is combined with a fallback
chain that leaves OpenAI.

### 5. The caller executes tools. The service never does.

```
app -> service -> provider     returns tool_calls
app executes the tool          its network, its credentials, its side effects
app -> service -> provider     with the tool result appended
```

The service stays stateless per request, holds no conversation, and never needs
credentials for a team's internal APIs. Server-side tool execution would be a
different product with a different security model.

### 6. Fallback is pinned once a tool has executed

```
turn 1  gpt-4o -> tool_calls: [charge_card(amount=500)]
        the app executes it; the customer's card is charged
turn 2  app returns the result; gpt-4o is rate-limited
```

If the fallback fires, Claude receives a history containing a tool call it never
made, may reason differently, and may emit `charge_card` again. The idempotency key
(Q15) protects **our** provider billing and does nothing for the app's side
effects, because the app is the one calling the payment API.

**Once a conversation contains a tool result, the provider is pinned for the rest
of the conversation and fallback is disabled.** A rate limit at that point is a
hard failure rather than a silent second charge.

This is consistent with ADR-0008: no failover once output is committed. An executed
tool call is a far stronger commitment than an emitted token.

Detection is cheap and needs no server-side state: if any message in the request
has role `tool`, the conversation is committed.

Teams whose tools are read-only can opt out per policy:

```yaml
policies:
  team-a-prod:
    tool_conversation_failover: pin     # default; "allow" is opt-in
```

### 7. Tools are contract class

Per ADR-0015 §3, a model that cannot do tools is removed from the fallback chain,
never sent a stripped request. A caller expecting `tool_calls` cannot use prose.

## Consequences

**Teams write one tool loop.** The same code runs against OpenAI, Anthropic, and a
self-hosted Llama, and survives a fallback between them.

**Argument validation is a guarantee we now owe.** If a model emits JSON that does
not match the declared schema, that is our error to classify and surface, not a
surprise in the caller's unmarshal. It also means a retry path for repairable
output is ours to decide later.

**Pinning costs availability on tool conversations.** A team doing multi-turn tool
work gets no fallback protection after the first tool result. That is the correct
trade against double-charging a customer, and it is visible in the response so they
can see when it happened.

**Cross-provider tool ids appear in history.** After a pin the ids stay consistent
by construction, but the `allow` opt-in means a conversation can carry ids in
another provider's format. Providers are permissive here today; if one tightens
validation, ids must be rewritten in the translation layer.

## Not yet ratified

- **`pin` as the default** is a recommendation, not an explicitly ratified call.
  The alternatives were: allow-and-mark (`provider_switched: true`, which most
  teams will not handle and which fails silently), and per-team policy only.
- **Parallel tool calls** are mappable in both directions
  (`parallel_tool_calls` / `disable_parallel_tool_use`) and are treated as tier-2
  normalization, but the behavior when a provider ignores the toggle is undecided.
