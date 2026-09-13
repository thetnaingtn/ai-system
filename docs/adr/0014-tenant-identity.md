# 14. Tenant identity is proven by the platform, resolved at the transport boundary

Date: 2026-09-06
Status: Accepted

## Context

Almost every decision already recorded needs to know which tenant a request came
from: per-team fallback lookup (ADR-0005), per-tenant usage counters and the
noisy-neighbor playbook (ADR-0003), `policy apply` authorization (ADR-0010), and
chargeback.

A tenant name carried in the request body is a claim, not an identity. Team B can
type `team-a` and spend team A's budget, read team A's policy, and overwrite team
A's config. It does not take malice — a copy-pasted manifest does it.

This is the one decision that cannot be deferred the way storage was (ADR-0013).
Adding authentication after clients exist breaks every client at once, on the
platform's schedule. The right time is now, at zero clients.

## Decision

### 1. Identity is resolved at the transport boundary, never from a proto field

A gRPC interceptor establishes the tenant before any handler runs. There is no
`tenant_id` field in any request message. If a handler needs the tenant it takes
it from the context.

```go
func TenantInterceptor(r TenantResolver) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, h grpc.UnaryHandler) (any, error) {
        t, err := r.Resolve(ctx)
        if err != nil {
            return nil, status.Error(codes.Unauthenticated, "no tenant identity")
        }
        return h(withTenant(ctx, t), req)
    }
}
```

### 2. Identity is proven by the platform, not issued by us

Kubernetes signs a token for every pod. The client mounts it; the model service
verifies the signature against the cluster's public keys. Neither service knows
the other exists — both trust the cluster.

Client side, once per service:

```yaml
volumes:
  - name: model-service-token
    projected:
      sources:
        - serviceAccountToken:
            path: token
            audience: model-service      # prevents replay to/from other services
            expirationSeconds: 3600
```

Platform side, once ever:

```yaml
auth:
  issuers:
    - url: https://kubernetes.default.svc
      audience: model-service
```

We issue no secrets, store no secrets, and rotate no secrets. Kubernetes rewrites
the token before expiry; the SDK re-reads the file rather than caching its
contents.

### 3. Verify with OIDC/JWKS, never with the Kubernetes TokenReview API

Both work for service account tokens. TokenReview is a Kubernetes API call on the
request path and makes the verification layer permanently Kubernetes-shaped.
JWKS is generic JWT verification: the same code accepts a GitHub Actions token, a
SPIFFE JWT-SVID, or a Cloud Run identity token by adding an issuer to a trust
list. Public keys are fetched once and cached.

Generic layer verifies signature, `iss`, `aud`, `exp`. Per-issuer resolvers read
their own claims.

### 4. Tenant comes from structured claims, not from parsing `sub`

A projected token carries structured Kubernetes claims. Read fields; do not do
string surgery on `sub`.

```go
type k8sClaims struct {
    Kubernetes struct {
        Namespace      string `json:"namespace"`
        ServiceAccount struct{ Name, UID string } `json:"serviceaccount"`
        Pod            struct{ Name, UID string } `json:"pod"`
    } `json:"kubernetes.io"`
}
```

`sub` is logged as a display string and never computed from. The pod name comes
free and is what answers "which replica is hammering us".

The namespace is used **verbatim** as the tenant id. No suffix trimming, no naming
convention to follow, no derivation that can guess wrong.

The value is validated before it becomes a tenant id, because it flows into
metric labels, config lookup keys, and eventually storage:

```go
var nsRE = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)  // DNS-1123 label

if ns == "" || len(ns) > 63 || !nsRE.MatchString(ns) {
    return Tenant{}, status.Error(codes.Unauthenticated, "malformed identity")
}
```

A malformed or missing claim is `Unauthenticated`. It is never a default tenant —
a parse failure that quietly attributes spend to the wrong place is worse than a
rejected request.

### 5. Record the fine identity, key on the coarse one

```go
type Tenant struct {
    ID             string // namespace, verbatim — policy and quota key on this
    Namespace      string
    ServiceAccount string // observability and future per-workload overrides
    PodName        string
}
```

Decisions key on `ID`. Logs and metrics carry the service account and pod. When a
team eventually asks for a per-workload quota, the data is already there: it is an
override table, not a wire-format change and not a client redeploy.

Explicit per-workload registration was rejected as the default: it charges
onboarding friction to every workload to solve a problem that will exist on a few.

### 6. Tenant is not team

A tenant is the unit spend is attributed to and policy is attached to. Sometimes
that is a team, sometimes one person's experiment, sometimes a platform component.
A namespace created by a single engineer is a tenant, with nothing missing. There
is no team concept in the identity path.

### 7. Unknown but authentic tenants are allowed, at the experimental tier

A verified token from a tenant we have never seen is served. It gets defaults: the
model named in the request, no fallback chain (ADR-0005), and the experimental
rate limit.

| tier | rate limit | quota | fallback | under pressure |
|---|---|---|---|---|
| `production` | high | tenant budget | as configured | protected |
| `experimental` (default) | low | small hard cap | none | shed first |

This inverts the risk correctly. Someone who never talked to us gets something
that works but cannot hurt a production path through the shared key (ADR-0003).
The conversation about raising limits is where we learn who they are — onboarding
is a consequence of success, not a gate on starting.

### 8. Ownership metadata is required to leave the experimental tier

```yaml
tenants:
  team-a-prod:
    owner: team-a-oncall@company.com
    cost_center: CC-4471
    tier: production
  ml-experiments:
    owner: someone@company.com      # a person is a valid owner
    cost_center: CC-0000
    tier: experimental
    review_by: 2027-03-01
```

Not authorization — metadata. But it is the only thing that answers who pays, who
to page, and whether an abandoned workload can be turned off. Unattributed spend is
a first-class number on the usage dashboard, and `review_by` is what kills orphans:
a monthly job flags expired experimental tenants, we ping the owner, silence means
it is switched off.

### 9. Principal, then tenant — one indirection

Resolvers produce a principal; a table maps principal to tenant.

```yaml
principals:
  "github:repo:company/team-a-app:ref:refs/heads/main": team-a-prod
  "gcp:checkout@team-a.iam.gserviceaccount.com":        team-a-prod
```

For Kubernetes the default rule (tenant = namespace) means the table starts empty.
It exists for two reasons:

- **Writes.** For `policy apply` this table *is* the authorization decision, so it
  lives in the platform repo under platform review. If a team could add its own
  row, team B adds one pointing at `team-a-prod`.
- **Portability.** A tenant id coined on Kubernetes outlives the platform that
  coined it. Without this, migrating means rewriting every policy key, breaking
  every dashboard, and severing billing history at the cutover — a data problem no
  code change fixes, and the reason "just swap the resolver" migrations go badly.

### 10. Reads self-onboard; writes require review

A new team calls `Generate` with no involvement from the platform team. The first
time they want a fallback chain, we review a PR. The gate is on writes and on
privileged tiers, not on getting started.

### 11. `policy apply` runs in CI, not on a laptop

Policy lives in the team's repo and reaches the store through their pipeline
(ADR-0010). Authorization is code review in their repo; the audit trail is git. A
human applying to production means the store and the repo can silently disagree.

CI is not a pod, so it authenticates by workflow OIDC — GitHub Actions mints a
token with `sub: repo:company/team-a-app:ref:refs/heads/main`, verified on the same
JWKS path, requiring the default branch. No secret is stored in CI. In-cluster CI
uses the existing pod path. A static token is the fallback only where no workflow
OIDC exists.

`platformctl policy apply` against production from a laptop is refused. Break-glass
during an incident is an explicitly named path (`--break-glass`), requires a human
token, logs loudly, and pages — because that need is real, and denying it means
someone reaches for `psql` instead.

### 12. Resolver interface, with a production fail-closed guard

```go
type TenantResolver interface {
    Resolve(ctx context.Context) (Tenant, error)
}
```

```
internal/auth/
  resolver.go     interface, chain resolver, JWKS verification, principal->tenant
  k8s.go          projected SA token           (v1, production)
  devheader.go    trusts x-tenant              (v1, local only)
  github.go       workflow OIDC for policy apply
  statictoken.go  CI and laptops, when needed
```

```go
if cfg.AuthMode == "insecure-header" && cfg.Env == "production" {
    log.Fatal("insecure-header auth is not permitted in production")
}
```

The SDK gets the mirror-image seam so a platform change is not a client redeploy:

```go
type CredentialSource interface {
    Token(ctx context.Context) (string, error)
}
```

with auto-detection — mounted SA token, then metadata server, then AWS credential
chain — so the default path needs no client configuration on any platform.

## Consequences

**Cost is near zero for readers.** A new team adds ~8 lines to a Deployment and
calls the SDK. No key is issued, no secret is handed over, nothing is rotated, no
meeting happens. This is the whole reason for choosing platform-signed identity
over static API keys.

**We must ship a copy-pasteable onboarding page.** Four steps, under a hundred
lines. If onboarding needs a meeting, the mechanism's advantage is wasted.

**The predictable failure is a mismatched audience**, which surfaces as a bare
`Unauthenticated`. The error must name both sides:

```
Unauthenticated: token audience ["default"] does not include "model-service".
Set audience: model-service in your serviceAccountToken volume. See <runbook>.
```

**Attribution stops at the namespace by default.** When a tenant's spend triples we
can say which tenant, not which of its services — from the policy engine's point of
view. The service account is in the logs and metrics, so the answer is available; it
is just not a key. Per-workload keying is an additive change when someone needs it.

**Migration cost is bounded, not eliminated.** With these seams: Cloud Run is a new
resolver (~1 day, same JWKS path); ECS is harder (~4 days) because AWS mints no
OIDC token for an arbitrary audience — it needs SigV4 plus `sts:GetCallerIdentity`,
or IAM Roles Anywhere, or SPIRE. Client teams do nothing in either case, and tenant
ids and billing history survive, because of decision 9.

**We are not running SPIRE.** SPIFFE/SPIRE is the stronger answer at scale —
per-workload certs, minutes-long TTL, platform-neutral — and it is a distributed
system we would have to operate for a service with zero clients today. JWKS
verification keeps the door open: a JWT-SVID is another issuer on the same path.

**Service mesh identity was rejected** for v1. `x-forwarded-client-cert` is only as
trustworthy as the guarantee that nothing can reach the pod except through its own
sidecar — an infrastructure invariant the service cannot verify itself. Adopting a
mesh to solve authentication is also a large dependency for a small problem.

## Alternatives considered

| Approach | Rejected because |
|---|---|
| Static API keys per team | We become the secret store: leaks, manual rotation, no expiry, a lookup on every request. A key stolen two years ago still works. |
| SPIFFE / SPIRE | Strongest option; an entire distributed system to operate before we have one client. Kept reachable via JWKS. |
| Service mesh header | Trust depends on an unverifiable network invariant; requires adopting a mesh. |
| Cloud IAM (IRSA / Workload Identity) | Cloud lock-in for something unrelated to the cloud; control-plane dependency on the request path. |
| External IdP (Okta, Auth0) | The client needs a client secret to get a token — the static-key problem one level removed — plus a SaaS on every internal call. Right for external partners. |
| Network policy / IP allowlist | Not identity. Anything landing in an allowed namespace inherits it. Defense in depth only. |
| Unverified `x-tenant` header | No authentication. Retained as the local-development resolver, with a production guard. |
| Explicit per-workload registration as the default | Charges registration friction to every workload for a problem that will exist on a few; a confusing deploy-time denial for teams that forget. Available later as an override. |
