package main

import (
	"path/filepath"
	"testing"
)

func newFixture(t *testing.T) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.json")
	docPath := filepath.Join(dir, "doc.md")
	mustWrite(t, filepath.Join(dir, "workflow.yml"), "name: test\n")
	mustWrite(t, manifestPath, fixtureManifest())
	return dir, manifestPath, docPath
}

func fixtureManifest() string {
	return `{"schema_version":"riido-provider-real-cli-observation.v1","id":"test","title":"Test","generated_doc":"doc.md","workflow":"workflow.yml","evidence_artifact":"artifact","providers":[{"id":"fake","display_name":"Fake","default_executable":"missing-riido-provider","override_env":"RIIDO_FAKE_PROVIDER_PATH","go_package":".","test_regex":"TestIntegration"}]}`
}
