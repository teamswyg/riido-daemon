# Workspace Invariants: Operations

[Back to invariants](../invariants.md)

Every operation has one responsibility, deterministic output, and persisted IR
events.

| Operation | Input | Output / IR event |
| --- | --- | --- |
| `PrepareWorkspace` | `taskID`, `runID`, repo ref, policy bundle, native config plan | workdir tree and `WorkdirCreated` |
| `MountRepo` | repo cache path, ref, isolation mode | code tree under workdir, with repo lock rules |
| `InjectNativeConfig` | policy bundle plus native config plan | config files and `NativeConfigInjected(files[], nativeConfigVersion)` |
| `RecordBaseline` | workdir state hash | persisted pre-run baseline |
| `CollectArtifacts` | run result | output, logs, and artifacts directories |
| `ArchiveWorkspace` | terminal run | `archive.json` and `WorkdirArchived(workdirPath, archiveURI)` |
| `CleanupWorkspace` | terminal run plus expired retention policy | actual filesystem removal |
| `ReinjectNativeConfig` | T-CONFIG trigger | native config refresh only and `ConfigTemplateReinjected(from, to)` |
