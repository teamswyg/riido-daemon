package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteContractLabRendersVisualEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.html")
	evidence := evidenceFile{Scenarios: []scenario{{
		ID:         "figma.onboarding",
		Status:     statusPassed,
		Screenshot: ".riido-local/screenshots/figma-onboarding.png",
	}}}
	if err := writeContractLab(path, evidence); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)
	for _, want := range []string{"visual evidence", "className: \"shot\"", "../screenshots/"} {
		if !strings.Contains(html, want) {
			t.Fatalf("contract lab missing %q", want)
		}
	}
}
