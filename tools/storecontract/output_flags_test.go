package main

import "testing"

func TestSelectedOutputPathPrefersEvidenceOut(t *testing.T) {
	got, err := selectedOutputPath("", "out/evidence.json")
	if err != nil {
		t.Fatal(err)
	}
	if got != "out/evidence.json" {
		t.Fatalf("path = %q", got)
	}
}

func TestSelectedOutputPathRejectsConflictingAliases(t *testing.T) {
	if _, err := selectedOutputPath("out/legacy.json", "out/evidence.json"); err == nil {
		t.Fatal("expected conflicting output aliases to fail")
	}
}
