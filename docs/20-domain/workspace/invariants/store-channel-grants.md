# Workspace Invariants: Store Channel Grants

[Back to invariants](../invariants.md)

Store channels separate workdir root from user workspace root.

| Root | Owner | Store rule |
| --- | --- | --- |
| `workdir root` | C11 app data root + C6 materialization | created inside app container, package local data, or app group |
| `user workspace root` | C11 WorkspaceGrantStore | inaccessible without user-selected folder grant |
| `repo cache` | C6, constrained by C11 path selection | arbitrary home scan is forbidden in store channel |

Rules:

1. macOS App Store requires a workspace grant expressed as security-scoped
   bookmark or app group/container grant. C6 consumes only the allowed root
   produced by C11.
2. Windows MSIX requires package identity and user-selected folder grant. C6 does
   not know picker or capability implementation details.
3. The current `dev-local` `~/Library/Application Support/riido` path cannot be
   shipped as a store artifact.
4. Provider cwd remains `workdir/`; selected workspace content is materialized by
   snapshot, worktree, or shallow clone during prepare.
5. C11 WorkspaceGrantStore is a later migration slice. C6 materializes user
   workspace root only when both active grant and
   `ConsentLedger.workspace-access:<workspace-id>` are true.
