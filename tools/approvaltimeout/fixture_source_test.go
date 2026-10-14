package main

const fixtureManifestSource = `{
  "schema_version": "riido-approval-wait-timeout.v1",
  "id": "provider-runtime-approval-wait-timeout",
  "title": "Approval Wait Timeout Ownership",
  "generated_doc": "docs/20-domain/provider-runtime/adapter-draft-fields/approval-wait-timeout.md",
  "evidence_artifact": "approval-wait-timeout-evidence",
  "sources": {
    "semantic_activity_manifest": "docs/20-domain/provider-runtime/adapter-draft-fields/idle-watchdog.riido.json",
    "provider_draft_manifest": "docs/20-domain/provider-runtime/runtime-responsibility/provider-event-draft.riido.json"
  },
  "approval_event": {"event_kind": "tool_approval_needed", "event_type": "ApprovalRequested"},
  "timeout_event": {"event_kind": "timeout", "result_status": "timeout", "cancel_command": "cancel_provider"},
  "source_checks": [
    {"name": "hard", "file": "internal/agentbridge/session/session_runner_timers.go", "contains": "time.NewTimer(r.cfg.HardTimeout)"},
    {"name": "idle", "file": "internal/agentbridge/session/session_runner_timers.go", "contains": "time.NewTimer(r.cfg.SemanticIdle)"},
    {"name": "semantic", "file": "internal/agentbridge/session/session_runner_emit.go", "contains": "ev.Kind.IsSemanticActivity()"},
    {"name": "approval-hard", "file": "internal/agentbridge/session/session_tool_approval_resolver.go", "contains": "case <-r.hardC:"},
    {"name": "approval-idle", "file": "internal/agentbridge/session/session_tool_approval_resolver.go", "contains": "case <-r.idleC:"}
  ]
}`

const fixtureSemanticSource = `{"semantic_activity":["tool_approval_needed"]}`

const fixtureDraftSource = `{
  "mapped_events":[{"event_kind":"tool_approval_needed","event_type":"ApprovalRequested"}],
  "skipped_events":[{"event_kind":"timeout","reason":"session actor"}]
}`
