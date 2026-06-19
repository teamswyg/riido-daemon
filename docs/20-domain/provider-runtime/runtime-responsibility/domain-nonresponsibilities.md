# Domain Non-Responsibilities

[Back to runtime-responsibility.md](../runtime-responsibility.md)

C4 Provider Runtime does not:

- append `CanonicalEvent` to the IR log
- create or update `ProviderCapability`
- lease / claim / heartbeat tasks
- write workdir or native config
- decide policy, sandbox, or protected paths
- judge validation results

Ownership:

| Concern | Owner |
| --- | --- |
| IR append authority | `riido-contracts` IR append authority + public daemon `internal/ir/ingest` |
| provider capability facts | C3 Provider Capability |
| task lease / claim / heartbeat | C5 Runtime Scheduling |
| workdir / native config | C6 Workspace |
| policy / sandbox / protected path | C7 Security / Policy |
| validation outcome | C8 Validation |
