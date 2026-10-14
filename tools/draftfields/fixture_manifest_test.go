package main

const fixtureManifestSource = `{
  "schema_version": "riido-draft-field-surface.v1",
  "id": "provider-runtime-draft-field-surface",
  "title": "Draft Field Surface",
  "allowed_doc": "docs/allowed-fields.md",
  "forbidden_doc": "docs/forbidden-fields.md",
  "evidence_artifact": "draft-fields-evidence",
  "forbidden_scope": ["internal/agentbridge/supervisor/provider_event_draft.go"],
  "allowed_fields": [
    {"field": "Type", "status": "implemented", "meaning": "type", "source": "internal/agentbridge/supervisor/provider_event_draft.go", "contains": "(ir.EventType, map[string]any, bool)"},
    {"field": "ProviderTurnID", "status": "reserved", "meaning": "turn"}
  ],
  "forbidden_fields": [
    {"field": "EventID", "filled_by": "ingest", "reason": "authority", "tokens": ["EventID"]}
  ],
  "assertions": ["adapter draft surface is observation-only"]
}`
