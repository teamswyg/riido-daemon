package main

const fixtureManifestSource = `{
  "schema_version": "riido-process-lifecycle.v1",
  "id": "provider-runtime-process-lifecycle",
  "title": "Process Lifecycle",
  "generated_doc": "docs/20-domain/provider-runtime/adapter-draft-fields/process-lifecycle.md",
  "evidence_artifact": "process-lifecycle-evidence",
  "steps": [{"step": "spawn", "responsibility": "spawn"}],
  "interfaces": [
    {"name": "Adapter", "file": "internal/agentbridge/adapter.go", "methods": ["BuildStart", "NewParser", "Translate"]},
    {"name": "Parser", "file": "internal/agentbridge/adapter_parser.go", "methods": ["FeedStdout", "FeedStderr", "Close"]},
    {"name": "Process", "file": "internal/process/port.go", "methods": ["Start"]},
    {"name": "RunningProcess", "file": "internal/process/port.go", "methods": ["Stdout", "Stderr", "Exited", "WriteStdin", "CloseStdin", "Kill"]}
  ],
  "source_checks": [
    {"name": "spawn", "file": "internal/agentbridge/session/start.go", "contains": "cfg.Process.Start(ctx, cfg.Spawn)"},
    {"name": "stdout", "file": "internal/agentbridge/session/session_runner_new.go", "contains": "stdoutCh:  proc.Stdout()"}
  ]
}`
