package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

const (
	featureUIDSLPath       = "../../docs/30-architecture/figma-ai-agent-daemon-boundary/feature-ui.dsl.json"
	featureUIGeneratedPath = "feature_ui.generated.json"
)

func TestFeatureUIDSLGeneratedFileFresh(t *testing.T) {
	dsl := readFeatureUIJSON(t, featureUIDSLPath)
	generated := readFeatureUIJSON(t, featureUIGeneratedPath)
	if !bytes.Equal(canonicalJSON(t, dsl), canonicalJSON(t, generated)) {
		t.Fatal("feature_ui.generated.json is stale; regenerate from feature-ui.dsl.json")
	}
}

func TestFeatureUIDesignArtifactsStaySmall(t *testing.T) {
	for _, path := range []string{featureUIDSLPath, featureUIGeneratedPath} {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if lines := strings.Count(string(body), "\n") + 1; lines > 75 {
			t.Fatalf("%s has %d lines, want <= 75", path, lines)
		}
	}
}

func TestFeatureUIScenarioExposesCaptureAndTolerance(t *testing.T) {
	got := featureUIScenario()
	if got.ID != "contract.ui.feature_dsl" || got.Screenshot == "" {
		t.Fatalf("feature ui scenario missing capture: %+v", got)
	}
	golden, ok := got.Observed["golden"].(map[string]any)
	if !ok || golden["max_pixels"] == nil || golden["max_ratio"] == nil {
		t.Fatalf("feature ui tolerance missing: %+v", got.Observed)
	}
}

func readFeatureUIJSON(t *testing.T, path string) any {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func canonicalJSON(t *testing.T, value any) []byte {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return body
}
