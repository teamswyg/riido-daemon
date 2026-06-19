# Import Rules

[Back to Module Decomposition SSOT](../module-decomposition.md)

| Package group | May import | Must not import |
| --- | --- | --- |
| `internal/agentbridge` root | stdlib | provider packages, process implementations, local API, task DB, mwsd/project, SaaS HTTP adapters |
| `internal/agentbridge/session` | `internal/agentbridge`, `internal/process` | concrete provider packages, task DB, mwsd/project, local API |
| `internal/provider/<name>` | `internal/agentbridge`, provider-local helpers, allowed policy/workdir helpers | another provider package, local task DB/project/server internals |
| `internal/scheduling` | stdlib and contract capability types | provider/process/local API/project/server implementations |
| `internal/hostintegration` | stdlib and contract vocabulary | provider execution, task DB/project, local API, workdir, server/deploy code |
| `internal/policy` | stdlib and C11 value types | provider execution, OS adapters, task DB/project, server/deploy code |
| `internal/riidoapi` | local task/validation adapters and local transports | public TCP listener, SaaS server packages |
| `cmd/riido` | composition packages in this repository | server binary code, Terraform/AWS/deploy evidence, provider CLI binaries |
| Future Store App GUI adapter | C11/local API contracts over IPC | daemon internals, provider execution packages, copied C11 domain facts, signing/provisioning secrets |

Production code must keep these rules. Tests may use package-local fakes, but
must not normalize cross-context imports as production design.
