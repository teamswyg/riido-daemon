package codex

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestModelReasoningEfforts(t *testing.T) {
	var model map[string]any
	if err := json.Unmarshal([]byte(`{
		"supportedReasoningEfforts": [
			{"reasoningEffort":"medium"}, {"reasoningEffort":"high"},
			{"reasoningEffort":"medium"}, {}, null
		],
		"defaultReasoningEffort":"high"
	}`), &model); err != nil {
		t.Fatal(err)
	}
	efforts, defaultEffort := modelReasoningEfforts(model)
	if !slices.Equal(efforts, []string{"medium", "high"}) || defaultEffort != "high" {
		t.Fatalf("got efforts %v and default %q", efforts, defaultEffort)
	}

	model["defaultReasoningEffort"] = "low"
	_, defaultEffort = modelReasoningEfforts(model)
	if defaultEffort != "" {
		t.Fatalf("unsupported default %q", defaultEffort)
	}
}
