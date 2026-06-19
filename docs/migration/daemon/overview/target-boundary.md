# Target Boundary

[Back to Overview](../overview.md)

Move into `riido-daemon`:

- provider-neutral runtime/session actors
- provider adapter ACLs for Claude, Codex, OpenClaw, and Cursor
- process spawning ports and fakes
- local-only daemon control surfaces
- host integration models for store-safe local execution
- daemon-side validation and black-box tests
- daemon SSOT docs and daemon-specific ADRs

Do not move:

- `cmd/riido_ai_server` or `internal/riidoaiserver`
- Terraform, AWS, ECS, ECR, WAF, ACM, Route53, or release evidence workflows
- `.riido-local`, state files, credentials, account IDs, or deploy artifacts
- shared contract code consumed by both daemon and control-plane
- bundled Claude/Codex/OpenClaw/Cursor CLI binaries
