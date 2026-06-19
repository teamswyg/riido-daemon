# Current Structure Evidence

[Back to Overview](overview.md)

| Observation | Current SSOT | Meaning |
| --- | --- | --- |
| SaaS assignment id is execution id | `saasplane/runtime_id.go` | same task can carry multiple assignments |
| state maps use execution id | `state_loop.go` | watcher/runtime/body cleanup is assignment-scoped |
| supervisor duplicate guard uses `TaskRequest.ID` | `supervisor.go` | SaaS adapter can supply assignment id as run key |
| heartbeat reports running ids | `saasplane/http_client.go` | active assignment refresh needs no task-id reverse lookup |
| workspace domain already has per-run workdir | `docs/20-domain/workspace.md` | DTO still needs structured repo/workspace plan |
| provider launch PATH is frozen | `runtimeactor`, `detectutil` | detect and spawn now share child tool lookup |
| C7 gate fails closed without approval path | `toolpolicy`, `session` | web approval DTO is the wider shared lifecycle step |
