package providervalidation

import "testing"

func TestProviderValidationManifestIdentity(t *testing.T) {
	manifest := loadProviderValidationContext(t).manifest
	if manifest.SchemaVersion != "riido-daemon-provider-validation-matrix.v1" {
		t.Fatalf("schema_version = %q", manifest.SchemaVersion)
	}
	if manifest.ID != "daemon-provider-validation-matrix" || manifest.RiidoTask != "RIID-4901" {
		t.Fatalf("manifest identity drifted: %+v", manifest)
	}
	if manifest.HumanDoc != "docs/30-architecture/integration-matrix.md" {
		t.Fatalf("human_doc = %q", manifest.HumanDoc)
	}
	for _, want := range []string{
		"docs/20-domain/security.md",
		"docs/20-domain/provider-runtime.md",
		"docs/30-architecture/integration-matrix.md",
	} {
		if !hasString(manifest.SourceDocuments, want) {
			t.Fatalf("source_documents must include %q: %+v", want, manifest.SourceDocuments)
		}
	}
	if !hasString(manifest.GlobalRules, "Provider full-access/trusted runtime modes must be explicit daemon-owned harness envelopes, never implicit provider defaults or caller-provided CustomArgs.") {
		t.Fatalf("global_rules must preserve full-access harness invariant: %+v", manifest.GlobalRules)
	}
	for _, want := range []string{
		"provider-validation-matrix/claude.riido.json",
		"provider-validation-matrix/codex.riido.json",
		"provider-validation-matrix/openclaw.riido.json",
		"provider-validation-matrix/cursor.riido.json",
	} {
		if !hasString(manifest.ProviderFiles, want) {
			t.Fatalf("provider_files must include %q: %+v", want, manifest.ProviderFiles)
		}
	}
}
