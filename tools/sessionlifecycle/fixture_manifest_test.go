package main

const fixtureManifestSource = `{
  "schema_version": "riido-session-lifecycle.v1",
  "id": "provider-runtime-session-lifecycle",
  "title": "Session Lifecycle",
  "generated_doc": "docs/session-lifecycle.md",
  "evidence_artifact": "session-lifecycle-evidence",
  "steps": [
    {"step": "pin", "status": "implemented", "responsibility": "pin", "source_checks": ["pin"]}
  ],
  "source_checks": [
    {"name": "pin", "file": "internal/agentbridge/supervisor/provider_event_draft.go", "contains": "EventSessionPinned"}
  ],
  "absent_surfaces": [
    {"name": "direct ResumeSession API", "scope": ["internal/agentbridge"], "tokens": ["ResumeSession("], "reason": "no direct api"}
  ],
  "assertions": ["pin is source checked"]
}`
