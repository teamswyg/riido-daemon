package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func writeFixtureManifest(t *testing.T, root string) string {
	t.Helper()
	path := filepath.Join(root, "manifest.json")
	mustWrite(t, path, `{
  "schema_version":"riido-daemon-self-improvement-evidence.v1",
  "id":"fixture",
  "title":"Fixture",
  "generated_doc":"`+filepath.Join(root, "self.md")+`",
  "workflow":".github/workflows/self-improvement-evidence.yml",
  "evidence_artifact":"self-improvement-evidence",
  "loop_source":"docs/30-architecture/loop-engineering/self-improvement-evidence.riido.json",
  "required_evidence":[{"id":"loop","file":"loop.json","description":"loop","producer_command":"go run ./tools/loopevidence -evidence-out <evidence-dir>/loop.json","assertions":[{"field":"status","equals":"verified"},{"field":"problem_count","equals":0}]}],
  "closed_loop_classes":[
    {"id":"bug","kind":"bug","description":"bug closure","evidence_ids":["loop"]},
    {"id":"feature","kind":"feature","description":"feature closure","evidence_ids":["loop"]}
  ]
}`)
	return path
}

func readReport(t *testing.T, path string) report {
	t.Helper()
	var out report
	body := mustRead(t, path)
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatal(err)
	}
	return out
}
