# Workspace Invariants: Lifecycle

[Back to invariants](../invariants.md)

Workspace lifecycle is domain-level. Implementation may remain simple operations.

| State | Meaning | Entry condition |
| --- | --- | --- |
| `WorkspaceUnprepared` | task exists but workdir does not | `TaskCreated` through `TaskQueued` |
| `WorkspacePreparing` | workdir creation / repo mount / native config injection is in progress | `TaskClaimed` to `WorkdirPreparing` |
| `WorkspacePrepared` | preparation is complete, `NativeConfigVersion` is fixed, provider may start | precondition for `RunStarted` |
| `WorkspaceDirty` | run is active or terminal artifacts remain | after `RunStarted`, before `ArchiveWorkspace` |
| `WorkspaceArchived` | artifact retention location is recorded | after task terminal |
| `WorkspaceFailed` | prepare or archive failed irrecoverably | any failed phase |

The lifecycle is not a 1:1 mapping to `TaskState` (C1). A failed task may leave
workspace in `WorkspaceDirty` for operator analysis before `WorkspaceArchived`.
