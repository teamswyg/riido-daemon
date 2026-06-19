# Workspace Invariants: Directory Layout

[Back to invariants](../invariants.md)

`RIIDO_WORKDIR_ROOT` owns the exact workdir root path. `dev-local` and
`developer-id` default to `$HOME/Library/Application Support/riido/workspaces`.
Store channel app data roots are owned by C11 Distribution / Host Integration;
C6 only materializes roots allowed by C11 and never creates store-channel home
fallbacks.

```txt
$RIIDO_WORKDIR_ROOT/
  {workspace_id}/
    tasks/
      {task_id}/
        runs/
          {run_id}/
            workdir/
            output/
            logs/
            artifacts/
            native-config/
            ir/
            archive.json
```

Rules:

1. Provider cwd is always `workdir/`. Sandbox policy prevents relative access to
   sibling directories such as `../output/`.
2. Protected status for `output/`, `logs/`, `artifacts/`, `native-config/`, and
   `ir/` is decided by C7 `T-PATH`.
3. Terminal local archive defaults to `keep-in-place`; `archiveURI` is
   `file://<run-root>`.
4. Daemon stop that cancels an in-flight run still follows terminal workspace
   lifecycle and writes `archive.json`.
5. Filesystem cleanup is disabled by default. `RIIDO_WORKDIR_RETENTION_SECONDS`
   deletes only `keep-in-place` run roots whose `archive.json.archived_at` is
   older than the cutoff.
6. Size-based or task-count cleanup is not a local daemon default.
