# Security Invariants: Policy Targets

[Back to invariants](../invariants.md)

Each target is a member of `AllowedSurfaceSet` and is decided independently by
trust tier.

| ID | Target | Policy expression |
| --- | --- | --- |
| T-PERM | permission mode | provider permission mode enum, such as Claude `default`, `acceptEdits`, `plan`; `bypassPermissions` rejected |
| T-SBX | sandbox mode | enum plus provider activation decision, such as `read-only`, `workspace-write`, policy-owned `danger-full-access` |
| T-NET | network egress | mode plus allowlist: default-off, explicit allowlist, or unrestricted |
| T-PATH | protected paths | path glob list such as `.git/**`, `.env*`, secrets dirs, prod config |
| T-SEC | secret exposure | scoped-token policy: TTL, scope limit, env delivery rejection, log redaction |
| T-DESTR | destructive command | blocked shell command patterns such as `rm -rf`, `dd of=/dev/`, force push, DB drop |
| T-PUSH | git push / deploy / migration | repo, branch, and env allow matrix; `main` push needs human approval |
| T-MCP | MCP server allowlist | allowed MCP server ids plus transport limits |
| T-CFG | native config injection | task-level policy files materialized into workdir |

Secret exposure redaction marker/catalog details are owned by
[`../security-redaction.md`](../security-redaction.md).

This document owns target enum values and meanings. Adjacent contexts consume
only decisions, such as "T-SBX -> workspace-write".
