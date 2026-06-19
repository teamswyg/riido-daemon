# CLI Surface SSOT

> Riido task: RIID-4714 `[Cli] Architecture SSOT docs migration`

This file owns the public `cmd/riido` command boundary. `cmd/riido` is a
local-only adapter shell for the customer-PC daemon and local task tooling.
The broader Figma AI Agent daemon boundary is projected in
[`figma-ai-agent-daemon-boundary.md`](figma-ai-agent-daemon-boundary.md).

## Detail Surfaces

- [Role](cli-surface/role.md)
- [Command groups](cli-surface/command-groups.md)
- [Runtime settings mapping](cli-surface/runtime-settings.md)
- [Local IPC rule](cli-surface/local-ipc-rule.md)
- [Guarded mutation rule](cli-surface/guarded-mutation-rule.md)
- [Validation](cli-surface/validation.md)
