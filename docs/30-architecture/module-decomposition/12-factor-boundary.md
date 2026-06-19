# 12-Factor Boundary

[Back to Module Decomposition SSOT](../module-decomposition.md)

Configuration is injected through `RIIDO_*` env vars or explicit local CLI
flags. Test-only gates use `AGENTBRIDGE_*`. Local daemon state is disposable;
durable facts live in task DB JSON, sidecar lease/registry files, mwsd
projection files, workdir metadata, or SaaS assignment events. `cmd/riido`
opens local IPC only and must not add a public HTTP listener.
