package taskdb

import (
	"path/filepath"
	"testing"
)

func TestGuardedEvidencePersistsDeterministicResultAndSaveLoad(t *testing.T) {
	db := queueSampleTask(t)
	updated, evidence, receipt, err := AddGuardedTaskEvidence(db, sampleEvidenceInput(), fixedTime().Add(testMinute))
	if err != nil {
		t.Fatalf("AddGuardedTaskEvidence returned error: %v", err)
	}
	if evidence.Result != "passed" || evidence.ValidationGate != TaskEvidenceValidationV1 || receipt.Result != "passed" {
		t.Fatalf("unexpected evidence/receipt: evidence=%+v receipt=%+v", evidence, receipt)
	}

	path := filepath.Join(t.TempDir(), "task-db.json")
	if err := SaveTaskDB(path, updated); err != nil {
		t.Fatalf("SaveTaskDB returned error: %v", err)
	}
	loaded, err := LoadTaskDB(path)
	if err != nil {
		t.Fatalf("LoadTaskDB returned error: %v", err)
	}
	if loaded.SchemaVersion != TaskDBSchemaVersion || len(loaded.Evidence) != 1 || len(loaded.CommandReceipts) != 2 {
		t.Fatalf("loaded task DB mismatch: %+v", loaded)
	}
}
