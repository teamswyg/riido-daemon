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

Move into the CLI slice:

- `riido mwsd ...`
- `riido task ...`
- `riido serve`
- `riido api ...`
- `riido daemon ...`
- `riido bridge ...`
- local smoke commands that exercise the CLI as a black box
- usage/help tests that keep `printUsage()` authoritative

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
| Task FSM legality | daemon domain package after migration |
| IR event schema | `riido-contracts` only after promotion |
| Provider process execution | daemon runtime packages |
| Local IPC transport | daemon host integration/local API packages |
| SaaS HTTP/SSE server | `riido-control-plane` |
| Deploy/apply behavior | `riido-infra` |

## Migration Order

1. Move CLI docs and README examples.
2. Move CLI parser/usage tests that do not need migrated internals.
3. Move task command wrappers once their backing packages move. The local API
   backing package moved in RIID-4684 and should be consumed rather than
   redefined by CLI code.
4. Restore smoke scripts as black-box tests.
5. Keep real provider CLI tests opt-in and skipped unless executables exist.

## Validation Gates

Required before a CLI migration PR is mergeable:

```bash
go test ./...
go list -m all
```

After the full CLI implementation migrates, restore these checks:

```bash
go build ./cmd/riido
go run ./cmd/riido --help
go run ./cmd/riido bridge providers
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
