package main

import (
	"strings"
	"testing"
)

func TestValidateManifestReportsShapeAndArtifactProblems(t *testing.T) {
	m := manifest{
		SchemaVersion: "bad",
		Providers: []provider{
			{ID: "same"},
			{ID: "same", DisplayName: "Same", DefaultExecutable: "same",
				OverrideEnv: "RIIDO_SAME", GoPackage: "missing.go", TestRegex: "Test"},
		},
	}
	problems := validateManifest(t.TempDir(), m)
	for _, want := range []string{
		"schema_version must be riido-provider-real-cli-observation.v1",
		"id, title, generated_doc, workflow, and evidence_artifact are required",
		"provider rows require id, display_name, executable, env, package, and test",
		"duplicate provider id \"same\"",
		"missing artifact \"missing.go\"",
	} {
		assertProviderProblem(t, problems, want)
	}
}

func TestJoinProblemsKeepsBulletEvidence(t *testing.T) {
	got := joinProblems([]string{"first", "second"})
	if got != "- first\n- second\n" {
		t.Fatalf("unexpected joined problems %q", got)
	}
}

func TestProviderObservedDoesNotInventUnknownProviderFacts(t *testing.T) {
	if got := providerObserved(provider{ID: "unknown"}, "tool"); got != nil {
		t.Fatalf("unknown provider observed %#v", got)
	}
	observed := cursorObserved("")
	auth, ok := observed["auth_preflight"].(cursorAuthPreflight)
	if !ok || auth.InteractiveLoggedIn {
		t.Fatalf("unexpected cursor auth observation %#v", observed)
	}
}

func TestLocalBackendUnavailablePatterns(t *testing.T) {
	for _, text := range []string{"failovererror", "provider ollama", "cooldown"} {
		if !localBackendUnavailable(text) {
			t.Fatalf("expected backend pattern %q", text)
		}
	}
	if localBackendUnavailable("ordinary provider failure") {
		t.Fatal("unexpected backend unavailable match")
	}
}

func assertProviderProblem(t *testing.T, problems []string, needle string) {
	t.Helper()
	for _, problem := range problems {
		if strings.Contains(problem, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}
