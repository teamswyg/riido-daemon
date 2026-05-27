# riido-daemon

Customer-PC daemon and CLI for Riido.

This repository is the public, store-reviewable daemon boundary. It will own the
local daemon process, CLI surface, provider runtime adapters, local IPC, host
integration, and daemon-side validation.

## Module

```text
github.com/teamswyg/riido-daemon
```

This module consumes shared public contracts from:

```text
github.com/teamswyg/riido-contracts v0.3.0
```

## Repository Boundary

This repository may contain:

- local daemon and CLI code
- provider runtime adapters
- local IPC and host integration code
- daemon-side validation and black-box tests
- public SSOT documents that describe daemon behavior

This repository must not contain:

- Terraform state, AWS account details, or deployment secrets
- production infrastructure configuration
- bundled Claude/Codex/OpenClaw/Cursor CLI binaries
- private environment files or release artifacts

## Verification

```bash
go test ./...
go list -m all
```

The public GitHub Actions workflow rejects non-Riido Go dependencies so daemon
verification can run outside the private repository billing pool while still
sharing versioned Riido contracts.

Useful local daemon smoke commands:

```bash
go run ./cmd/riido --help
go run ./cmd/riido bridge providers
go run ./cmd/riido daemon start --socket /tmp/riido-agentd.sock --pid-file /tmp/riido-agentd.pid --log-file /tmp/riido-agentd.log
go run ./cmd/riido daemon status --socket /tmp/riido-agentd.sock
go run ./cmd/riido daemon stop --socket /tmp/riido-agentd.sock --pid-file /tmp/riido-agentd.pid
```

`riido daemon ...` chooses its task source through 12-factor environment:
`RIIDO_TASK_QUEUE_DIR`, `RIIDO_TASK_DB_SOURCE_PATH`, or `RIIDO_SAAS_URL` plus
`RIIDO_SAAS_AGENTS`. Provider CLIs remain external tools; this repository does
not bundle or install Claude, Codex, OpenClaw, or Cursor.

## License

Apache-2.0.
