# Provider Runtime / Adapter SSOT: Integration Gates

[Back to provider-runtime.md](../provider-runtime.md)

This entrypoint keeps grep-compatible markers while the evidence is split by
gate, adapter scope, and runtime snapshot semantics.

## Compatibility Markers

- `provider-validation-matrix.riido.json`
- `supports_worktree=false`
- `required_surfaces=[worktree]`
- `MISSING_REQUIRED_SURFACE:worktree`
- `runtime-snapshot`
- `RIID-4662`
- `RIID-4683`
- `RIID-4684`
- `RIID-4686`
- `RIID-4689`
- `RIID-4690`
- `RIID-4901`
- `A-54`

## Parts

- [Cursor real CLI gate](integration-gates/cursor-real-cli.md)
- [Provider validation matrix](integration-gates/provider-validation-matrix.md)
- [Supervisor migration scope](integration-gates/supervisor-scope.md)
- [Task DB plane migration scope](integration-gates/taskdbplane-scope.md)
- [Local API adapter scope](integration-gates/riidoapi-local-api.md)
- [MWSD/project sync scope](integration-gates/mwsd-project-sync.md)
- [SaaS plane migration scope](integration-gates/saasplane-scope.md)
- [SaaS runtime snapshot semantics](integration-gates/saasplane-runtime-snapshot.md)
- [Daemon lifecycle adapter scope](integration-gates/daemon-lifecycle-adapter.md)
