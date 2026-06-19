package main

const fixtureManifestSource = `{
  "schema_version": "riido-work-branch-gate.v1",
  "id": "riido-work-branch-gate",
  "title": "Riido Work Branch Gate",
  "generated_doc": "docs/branch.md",
  "generated_script": "scripts/verify-branch.sh",
  "workflow": ".github/workflows/branch.yml",
  "evidence_workflow": ".github/workflows/branch-evidence.yml",
  "evidence_artifact": "branch-gate-evidence",
  "pattern": "^[A-Z][A-Z0-9]*-[0-9]+-.+$",
  "shape": "<PROJECT_KEY>-<NUMBER>-<SLUG>",
  "example": "A-40-example",
  "allow_main": true,
  "rules": ["use Riido branchName"],
  "examples": [
    {"branch": "main", "accepted": true, "reason": "main"},
    {"branch": "A-40-good", "accepted": true, "reason": "shape"},
    {"branch": "codex/bad", "accepted": false, "reason": "namespace"}
  ],
  "assertions": ["script generated"]
}`
