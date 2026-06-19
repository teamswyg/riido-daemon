package main

const fixtureManifestSource = `{
  "schema_version": "riido-compatibility-gate.v1",
  "id": "test",
  "title": "Test Gate",
  "generated_doc": "docs/gate.md",
  "workflow": ".github/workflows/gate.yml",
  "evidence_artifact": "gate-evidence",
  "purpose": "test gate",
  "inputs": [{"name": "input", "owner": "owner", "source_checks": ["source"]}],
  "gate_order": [{"step": "step", "summary": "summary", "source_checks": ["source"]}],
  "outputs": ["output"],
  "failure_semantics": [{"case": "case", "meaning": "meaning", "source_checks": ["source"]}],
  "source_checks": [{"name": "source", "file": "internal/source.go", "contains": "needle"}],
  "assertions": ["all rows are checked"]
}`
