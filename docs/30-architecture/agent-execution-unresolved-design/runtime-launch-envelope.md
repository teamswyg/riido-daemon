# Runtime Launch Envelope

[Back to Overview](overview.md)

Provider detection and process spawn are separate phases. The launch envelope
keeps their shared assumptions explicit.

| Field | Owner | Rule |
| --- | --- | --- |
| `selected_executable` | daemon C4 | adapter-selected absolute executable path |
| `path` | daemon C4 | login-shell PATH plus known install dirs and policy env |
| `env` | daemon C4/C7 | sanitized child process env |
| `cwd` | daemon C6 | prepared workdir |
| `toolchain_probe` | daemon C4 | `git`, `node`, package manager availability |
| `detected_at` / `ttl` | daemon C4 | detection freshness |
| `approval_policy` | daemon C7 + contracts | auto-approve, approval-required, fail-closed surfaces |

Only fields needed by control-plane/client should be promoted immediately, such
as `selected_executable`, `provider_version`, `toolchain_probe`, and
`approval_policy_id`.
