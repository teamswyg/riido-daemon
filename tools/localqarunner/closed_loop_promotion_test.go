package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadClosedLoopPromotions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "promotions.json")
	body := `{
	  "schema_version": "riido-local-qa-closed-loop-promotions.v1",
	  "id": "local-qa-closed-loop-promotions",
	  "promotions": [{"candidate_id":"coverage.local-qa-daily-freshness"}]
	}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := loadClosedLoopPromotions(path)
	if len(got) != 1 || got[0].CandidateID != "coverage.local-qa-daily-freshness" {
		t.Fatalf("promotions=%+v", got)
	}
}

func TestLoadClosedLoopPromotionsIgnoresUnknownSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "promotions.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := loadClosedLoopPromotions(path); got != nil {
		t.Fatalf("expected nil promotions, got %+v", got)
	}
}
