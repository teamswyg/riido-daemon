package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

const (
	qaSystemDSLPath       = "../../docs/30-architecture/qa-system.dsl.json"
	qaSystemGeneratedPath = "qa_system.generated.json"
)

func TestQASystemDSLGeneratedFileFresh(t *testing.T) {
	dsl := readFeatureUIJSON(t, qaSystemDSLPath)
	generated := readFeatureUIJSON(t, qaSystemGeneratedPath)
	if !bytes.Equal(canonicalJSON(t, dsl), canonicalJSON(t, generated)) {
		t.Fatal("qa_system.generated.json is stale; run go generate ./tools/localproductacceptance")
	}
}

func TestQASystemDesignArtifactsStaySmall(t *testing.T) {
	for _, path := range []string{qaSystemDSLPath, qaSystemGeneratedPath} {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if lines := strings.Count(string(body), "\n") + 1; lines > 75 {
			t.Fatalf("%s has %d lines, want <= 75", path, lines)
		}
	}
}

func TestQASystemScenarioAuditsChangeDetection(t *testing.T) {
	got := qaSystemScenario()
	if got.ID != "local.qa.dsl_system_audit" || got.Status != statusPassed {
		t.Fatalf("QA system audit failed: %+v", got)
	}
	if got.Observed["search_entries"] != 6 {
		t.Fatalf("unexpected QA system audit evidence: %+v", got.Observed)
	}
	inference, ok := got.Observed["inference_removed"].(map[string]any)
	if !ok || inference["system_reports_problems"] != true || inference["fully_systematized"] != true {
		t.Fatalf("inference audit missing: %+v", got.Observed["inference_removed"])
	}
	if inference["remaining_source_count"] != 0 {
		t.Fatalf("remaining source-only DSL count = %+v", inference)
	}
	checks, ok := got.Observed["change_detection"].([]map[string]any)
	if !ok || len(checks) < 6 {
		t.Fatalf("change detection evidence missing: %+v", got.Observed["change_detection"])
	}
	generated, ok := got.Observed["generated_checks"].([]map[string]any)
	if !ok || len(generated) != 6 {
		t.Fatalf("generated freshness evidence missing: %+v", got.Observed["generated_checks"])
	}
	for _, check := range generated {
		if check["status"] != statusPassed {
			t.Fatalf("generated check was not passed: %+v", check)
		}
	}
}
