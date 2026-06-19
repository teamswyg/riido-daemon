# Integration Assertions

[Back to provider integration matrix](../integration-matrix.md)

| Provider | Integration assertion |
| --- | --- |
| Claude | stream JSON flow reaches `ResultCompleted`, and the run writes the expected file artifact inside the daemon-selected workdir |
| Codex | app-server JSON-RPC initialize/thread/turn flow reaches `ResultCompleted`, launch shape is explicit `--sandbox danger-full-access`, and the run writes the expected file artifact inside the daemon-selected workdir |
| OpenClaw | JSON or NDJSON result flow reaches `ResultCompleted` with deterministic session id and uses the executable path that passed Detect. Optional artifact integration may pass in a preconfigured local OpenClaw environment, but SaaS completion alone must not be treated as filesystem side-effect evidence, and runtime routing remains `supports_worktree=false`. |
| Cursor | selected launch profile adds daemon-workdir `--trust` without `--yolo`, stream JSON flow reaches `ResultCompleted`, and the run writes the expected file artifact inside the daemon-selected workdir; missing login probe skips |
