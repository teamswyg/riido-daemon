# Workspace Invariants: Repo Cache Isolation

[Back to invariants](../invariants.md)

Premise: code-modifying task workdirs and source repo caches are different
directories.

Isolation modes:

| Mode | Behavior | Cost | Recommended when |
| --- | --- | --- | --- |
| git worktree | `git worktree add` from shared bare repo under `cache/repos/{repo_hash}` | saves disk and prepares quickly | many tasks use the same repo |
| shallow clone | independent per-task clone from cache | more disk and slightly slower prepare | stronger isolation is required |

Selection is determined by task definition plus C7 policy bundle. This document
owns only mode semantics.

Cache update policy:

- `cache/repos/{repo_hash}` is shared across tasks and needs short locks only
  for fetch/prune updates.
- Active task worktrees are not affected by later cache updates.
- Automatic shared cache prune is not a local daemon default. Operator-triggered
  prune must run under `repo_cache_update.lock` and must never delete active
  task workdirs or run roots.
