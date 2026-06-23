package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

const (
	qaI18NDSLPath       = "../../docs/30-architecture/qa-i18n.dsl.json"
	qaI18NGeneratedPath = "qa_i18n.generated.json"
)

func TestQAI18NDSLGeneratedFileFresh(t *testing.T) {
	dsl := readFeatureUIJSON(t, qaI18NDSLPath)
	generated := readFeatureUIJSON(t, qaI18NGeneratedPath)
	if !bytes.Equal(canonicalJSON(t, dsl), canonicalJSON(t, generated)) {
		t.Fatal("qa_i18n.generated.json is stale; run go generate ./tools/localproductacceptance")
	}
}

func TestQAI18NDesignArtifactsStaySmall(t *testing.T) {
	for _, path := range []string{qaI18NDSLPath, qaI18NGeneratedPath} {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if lines := strings.Count(string(body), "\n") + 1; lines > 75 {
			t.Fatalf("%s has %d lines, want <= 75", path, lines)
		}
	}
}

func TestQAI18NScenarioDefaultsToKorean(t *testing.T) {
	got := qaI18NScenario()
	if got.ID != "contract.ui.i18n_dsl" || got.Status != statusPassed {
		t.Fatalf("i18n scenario missing: %+v", got)
	}
	if got.Observed["default_locale"] != "ko" || got.Observed["fallback_locale"] != "en" {
		t.Fatalf("unexpected locale config: %+v", got.Observed)
	}
	if !qaI18NHasMessage(t, got.Observed, "domain", "title", "내부 실행 확인 화면") {
		t.Fatalf("domain title Korean message missing: %+v", got.Observed)
	}
}

func TestQAI18NTranslationCoverageComplete(t *testing.T) {
	got := qaI18NScenario()
	coverage, ok := got.Observed["translation_coverage"].(map[string]any)
	if !ok {
		t.Fatalf("translation coverage missing: %+v", got.Observed)
	}
	if coverage["passed"] != true {
		t.Fatalf("translation coverage failed: %+v", coverage)
	}
	if coverage["namespace_count"] != 16 || coverage["message_count"] != 166 || coverage["required_cell_count"] != 332 {
		t.Fatalf("unexpected translation scope: %+v", coverage)
	}
	if coverage["translated_cell_count"] != 332 || coverage["missing_cell_count"] != 0 {
		t.Fatalf("unexpected translation coverage: %+v", coverage)
	}
	if coverage["locale_error_count"] != 0 || coverage["placeholder_mismatch_count"] != 0 || coverage["duplicate_key_count"] != 0 {
		t.Fatalf("translation integrity failed: %+v", coverage)
	}
	if coverage["coverage_ratio"] != 1.0 {
		t.Fatalf("coverage ratio = %+v", coverage["coverage_ratio"])
	}
}

func qaI18NHasMessage(t *testing.T, spec map[string]any, namespace, key, ko string) bool {
	t.Helper()
	for _, rawNS := range spec["namespaces"].([]any) {
		ns := rawNS.(map[string]any)
		if ns["id"] != namespace {
			continue
		}
		for _, rawMessage := range ns["messages"].([]any) {
			message := rawMessage.([]any)
			if message[0] == key && message[1] == ko {
				return true
			}
		}
	}
	return false
}
