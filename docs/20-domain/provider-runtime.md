# Provider Runtime / Adapter SSOT

This SSOT is split into focused topic files so each reader can open the smallest relevant surface.
The original entrypoint remains here to preserve existing links.

## Compatibility Markers

- `figma-ai-agent-daemon-boundary`
- `RIID-4901`
- `provider-validation-matrix.riido.json`
- `supports_worktree=false`
- `required_surfaces=[worktree]`
- `MISSING_REQUIRED_SURFACE:worktree`
- Provider full-access/trusted modes are not assumed from provider defaults or
caller arguments
- daemon-owned full-access runtime selection
- Codex adapter 가 danger-full-access envelope 만 생성하고 그 위험을 Riido harness 가
관리한다
- not a provider default, caller-provided default, or
  hidden fallback
- Other providers should follow the same full-access/trusted-runtime
meta model only through provider-specific SSOT
- `internal/process.DefaultStdoutBuffer`
- `internal/process.DefaultStderrBuffer`
- `internal/agentbridge/session.DefaultEventBuffer`
- `internal/agentbridge/session.DefaultResultBuffer`
- `internal/agentbridge/runtimeactor.DefaultMailboxSize`
- `internal/agentbridge/supervisor.DefaultMailboxSize`
- `Q-RT-001` closes alongside `Q-MULTICA-005`
- `Q-CTX-001`
- `Q-RT-003` closed
- C4 Provider Runtime / Adapter owns approval wait
- EventIngestor does not own approval

## Parts

- [Overview](provider-runtime/overview.md)
- [Public Migration Status](provider-runtime/public-migration-status.md)
- [Integration Gates](provider-runtime/integration-gates.md)
- [Runtime Responsibility](provider-runtime/runtime-responsibility.md)
- [Adapter Draft Fields](provider-runtime/adapter-draft-fields.md)
- [Adapter ACL](provider-runtime/adapter-acl.md)
- [RuntimeActor Boundary](provider-runtime/runtime-actor-boundary.md)
- [Versioning](provider-runtime/versioning.md)
