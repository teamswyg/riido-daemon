package main

const fixtureManifestSource = `{
  "schema_version": "riido-saas-assignment-source.v1",
  "id": "test",
  "title": "SaaS Assignment",
  "generated_doc": "docs/domain.md",
  "migration_doc": "docs/migration.md",
  "workflow": ".github/workflows/saas-assignment-source.yml",
  "evidence_artifact": "saas-assignment-source-evidence",
  "purpose": "test",
  "facts": [{"name": "fact", "status": "implemented", "summary": "fact", "source_checks": ["source"]}],
  "boundaries": [{"name": "boundary", "owner": "owner", "summary": "summary"}],
  "absent_surfaces": [{"name": "forbidden", "scope": ["internal"], "tokens": ["thread-progress"], "reason": "not owned"}],
  "source_checks": [{"name": "source", "file": "internal/source.go", "contains": "needle"}],
  "assertions": ["implemented facts have source evidence"]
}`
