package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteEvidenceCreatesReadableJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "evidence.json")
	evidence := evidenceFile{
		SchemaVersion: "riido-product-acceptance.v1",
		ID:            "ai-agent-product-acceptance",
		ObservedAt:    time.Unix(1, 0).UTC().Format(time.RFC3339),
		ExpiresAt:     time.Unix(2, 0).UTC().Format(time.RFC3339),
		Status:        statusPartial,
		Scenarios:     []scenario{{ID: "contract.task.sse_replay", Status: statusFailed}},
	}
	if err := writeEvidence(path, evidence); err != nil {
		t.Fatalf("write evidence: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Fatalf("evidence should be newline-terminated: %q", data)
	}
	var decoded evidenceFile
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode evidence: %v", err)
	}
	if decoded.Status != statusPartial || len(decoded.Scenarios) != 1 {
		t.Fatalf("unexpected decoded evidence: %#v", decoded)
	}
}

func TestContractUIScenarioDocumentsFrontendHandoff(t *testing.T) {
	sc := contractUIScenario("/tmp/lab.html", "/tmp/manual.json")
	if sc.ID != "contract.ui.lab" || sc.Status != statusPassed {
		t.Fatalf("unexpected contract UI scenario: %#v", sc)
	}
	if sc.Observed["client_mutations"] != false {
		t.Fatalf("client mutations must stay false: %#v", sc.Observed)
	}
	if sc.Observed["manual_evidence"] != "/tmp/manual.json" {
		t.Fatalf("manual evidence path missing: %#v", sc.Observed)
	}
}
