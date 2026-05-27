# Riido CLI Migration Plan

> Riido task: RIID-4635 `[Cli] 기존 riido_daemon CLI 마이그레이션 계획/문서화`

This document defines how the CLI surface of the former private `riido_daemon`
repository moves into the public `riido-daemon` repository.

## Goal

The CLI remains in `riido-daemon` because it is the user/operator entrypoint for
the customer-PC daemon. It should be documented separately from the long-running
daemon runtime so CLI command ownership, compatibility, and store-review
constraints stay clear.

## Source In The Private Repository

The initial CLI source is:

- `cmd/riido`
- CLI usage text owned by `printUsage()`
- CLI tests under `cmd/riido`
- scripts that call local CLI commands, when they are not deploy/infra scripts
- README command examples that describe local daemon and local API usage
- `docs/30-architecture/config-reference.md`
- `docs/20-domain/runtime-versioning.md`
- CLI-related rows in roadmap/audit docs

## Target Boundary

Move into the final CLI boundary:

- `riido mwsd ...`
- `riido task ...`
- `riido serve`
- `riido api ...`
- `riido daemon ...`
- `riido bridge ...`
- local smoke commands that exercise the CLI as a black box
- usage/help tests that keep `printUsage()` authoritative

RIID-4685 is a smaller public-safe slice inside that final boundary. It moves
only the commands whose backing packages are already public:

- `riido task ...` over `internal/taskdb`
- `riido serve` and `riido api ...` over `internal/riidoapi`
- `riido bridge providers|detect` over public provider adapters

`riido mwsd ...` remains deferred until `mwsdbridge/project` projection sync is
split from private workspace state. Full `riido daemon ...` process lifecycle
commands remain deferred until the daemon runtime wrapper no longer imports the
private SaaS plane or private projection source.

Do not move into the CLI slice:

- SaaS server binary `cmd/riido_ai_server`
- Terraform deploy commands or AWS apply workflows
- provider CLI binaries
- private environment files or machine-local state
- shared contract facts that belong in `riido-contracts`

## CLI / Daemon Split

The CLI is a thin adapter. Domain decisions must stay in the owning packages and
SSOT docs.

| CLI concern | Owner |
| --- | --- |
| Argument parsing and usage text | `cmd/riido` |
| Task FSM legality | `riido-contracts/task` through `internal/taskdb` guarded mutation |
| IR event schema | `riido-contracts/ir` |
| Provider process execution | daemon runtime packages |
| Local IPC transport | `internal/riidoapi` and `internal/hostintegration` |
| SaaS HTTP/SSE server | `riido-control-plane` |
| Deploy/apply behavior | `riido-infra` |

## Migration Order

1. Move CLI docs and README examples.
2. Move CLI parser/usage tests that do not need migrated internals.
3. Move task command wrappers once their backing packages move. RIID-4685 moves
   the task/API/bridge command wrappers against public `internal/taskdb`,
   `internal/riidoapi`, and provider adapter ports.
4. Restore smoke scripts as black-box tests.
5. Keep real provider CLI tests opt-in and skipped unless executables exist.

## Validation Gates

Required before a CLI migration PR is mergeable:

```bash
go test ./...
go list -m all
go build ./cmd/riido
go run ./cmd/riido --help
go run ./cmd/riido bridge providers
```

After the full CLI implementation migrates, restore these checks:

```bash
go run ./cmd/riido daemon status --socket /tmp/riido-agentd.sock
```

Commands that need a running daemon, mwsd socket, provider CLI, or local app
state should be black-box smoke tests with explicit skip conditions.

## Store Review Invariants

- The CLI must not create public network listeners.
- Provider CLIs are discovered external tools, not bundled payloads.
- Commands that mutate guarded task state must preserve approval-id and receipt
  rules.
- Unsafe provider flags must remain policy-gated.

## Open Follow-Ups

| Follow-up | Repository |
| --- | --- |
| Move daemon runtime backing packages. | `riido-daemon` / RIID-4636 |
| Promote shared DTOs only when needed by multiple repos. | `riido-contracts` / RIID-4637 |
| Keep SaaS server commands out of the local CLI binary. | `riido-control-plane` / RIID-4638 |
| Keep deploy/apply commands in private automation. | `riido-infra` / RIID-4639 |
