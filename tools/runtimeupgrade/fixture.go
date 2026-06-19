package main

const fixtureManifestSource = `{
  "schema_version": "riido-runtime-upgrade-flow.v1",
  "id": "test",
  "title": "Runtime Upgrade",
  "generated_doc": "docs/runtime-upgrade.md",
  "workflow": ".github/workflows/runtime-upgrade.yml",
  "evidence_artifact": "runtime-upgrade-evidence",
  "invariant": "no silent upgrade",
  "inputs": [{"name": "input", "status": "implemented", "owner": "owner", "source_checks": ["source"]}],
  "flow": [{"step": "reserved step", "status": "reserved", "summary": "reserved", "required_evidence": "test"}],
  "policies": [],
  "native_config": [],
  "source_checks": [{"name": "source", "file": "internal/source.go", "contains": "needle"}],
  "assertions": ["implemented rows have source checks"]
}`
