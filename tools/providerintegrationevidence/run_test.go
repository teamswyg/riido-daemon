package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCurrentManifestAndGeneratedDoc(t *testing.T) {
	if err := run("../..", "docs/30-architecture/provider-real-cli-observation.riido.json", "", false, true, false); err != nil {
		t.Fatal(err)
	}
}

func TestEvidenceOutputRecordsSkippedProviders(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.json")
	docPath := filepath.Join(dir, "doc.md")
	evidencePath := filepath.Join(dir, "evidence.json")
	data := `{
	  "schema_version":"riido-provider-real-cli-observation.v1",
	  "id":"test",
	  "title":"Test",
	  "generated_doc":"doc.md",
	  "workflow":"workflow.yml",
	  "evidence_artifact":"artifact",
	  "providers":[{"id":"missing","display_name":"Missing","default_executable":"definitely-missing-riido-provider","override_env":"RIIDO_MISSING_PATH","go_package":".","test_regex":"TestIntegration"}]
	}`
	mustWrite(t, manifestPath, data)
	mustWrite(t, filepath.Join(dir, "workflow.yml"), "name: test\n")
	mustWrite(t, docPath, renderMarkdown(mustLoad(t, manifestPath)))
	if err := run(dir, manifestPath, evidencePath, false, true, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(evidencePath); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustLoad(t *testing.T, path string) manifest {
	t.Helper()
	out, err := loadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
