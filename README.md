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

## License

Apache-2.0.
