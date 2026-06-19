# Workspace Plan

[Back to Overview](overview.md)

F3 is fixed by passing structured assignment snapshot fields, not by telling the
provider about a repo inside prompt text.

| Field | Phase | Rule |
| --- | --- | --- |
| `source_kind` | P0 | `empty`, `git_public`, `git_private_unsupported`, `git_private_token_ref` |
| `repo_url` | P0 public | public clone URL only |
| `repo_full_name` | P0 | stable display/diagnostic id |
| `branch_name` | P0 | Riido task branch or selected branch |
| `commit_sha` | P1 | exact checkout when present |
| `visibility` | P0 | `public`, `private`, `unknown` |
| `auth_mode` | P0 | `none`, `unsupported`, `token_ref` |
| `auth_ref` | P2 private | secret broker reference, never raw token |
| `isolation_mode` | P0 | `git_worktree`, `shallow_clone`, `empty_explicit` |
| `required_surfaces` | P0 | worktree/session/tool surfaces |

P0 clones public repositories only. Private repositories fail closed as
`git_private_unsupported` until token-ref broker ownership is defined.
