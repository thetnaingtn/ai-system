# Shared protobuf contracts

All platform `.proto` sources belong here, grouped by platform, service, and API
version. The initial contract lives in `proto/platform/model/v1/` with package
`platform.model.v1` and service `ModelService`.

| File | Contents |
|---|---|
| `platform/model/v1/model.proto` | Model metadata, authorized configuration, typed capabilities, and enums |
| `platform/model/v1/model_change.proto` | Pending/reviewed configuration changes |
| `platform/model/v1/model_service.proto` | Nine model-management RPCs, request/response messages, and error details |

Generated Go files are committed under `gen/go/platform/model/v1/` in the root
module, `github.com/thetnaingtn/ai-system`. Import them as
`github.com/thetnaingtn/ai-system/gen/go/platform/model/v1`, package `modelv1`.
Do not edit generated files.

## Toolchain and generation

Use the installed Buf CLI version 1.75.0. `buf.yaml` uses v2 configuration with
`proto/` as a single module, `STANDARD` linting, and `FILE` compatibility checks.
`buf.gen.yaml` invokes local Go tools rather than remote plugins or installed
standalone generator binaries:

```sh
make proto-tools
make proto-format
make proto-generate
make proto-check
```

`proto-tools` checks Buf and runs the generators' version commands, downloading
their modules through Go if needed. `go.mod` pins protoc-gen-go v1.36.11 and
protoc-gen-go-grpc v1.6.2 using `tool` directives. `go tool` builds/runs them from
Go's cache. There is no plugin installation into `bin/` or a global `GOBIN`.

`proto-check` builds/lints schemas, checks formatting, and compares committed Go
files with freshly generated files in a temporary directory. It detects missing,
stale, or extra generated files without rewriting the working tree.

There are no external protobuf schema dependencies yet; only bundled well-known
types are imported. No `buf.lock` is needed until external dependencies are added.
`go.work` continues to manage Go modules independently of the Buf workspace.

## Compatibility

After the first contract baseline exists on main:

```sh
make proto-breaking
```

The default baseline is `.git#branch=main`. It can be overridden with another
Git input or a previously built Buf image:

```sh
make proto-breaking BUF_BREAKING_AGAINST='.git#ref=<commit>'
```

The first contract addition has no prior Buf configuration on main and establishes
the baseline. CI runs only on pull requests and checks subsequent changes against
the PR's base SHA. Checks are local; the workflow does not publish schemas,
post PR comments, or require Buf registry credentials. Compatible schema changes
still require review for behavioral and authorization changes.

These definitions do not enable handlers on the model process, which remains
health-only. See the [model API design](../docs/ai-platform-model-api.md) and
[ADR-0034](../docs/adr/0034-buf-and-model-management-contract.md) for behavior and
implementation prerequisites.
