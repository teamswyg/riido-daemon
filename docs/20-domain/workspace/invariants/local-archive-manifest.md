# Workspace Invariants: Local Archive Manifest

[Back to invariants](../invariants.md)

The local filesystem adapter writes `riido-workdir-archive.v1` after
`ArchiveWorkspace` observes a terminal result. The manifest is written atomically
to the run root.

| Field | Meaning |
| --- | --- |
| `schema_version` | always `riido-workdir-archive.v1` |
| `workdir_path` | absolute path that was provider cwd |
| `archive_uri` | local default is `file://<run-root>` |
| `retention_mode` | local default is `keep-in-place` |
| `result_status` | terminal result: `completed`, `failed`, `cancelled`, `timeout`, `aborted`, or `blocked` |
| `archived_at` | archive write time; not part of `NativeConfigVersion` input |
