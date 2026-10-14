package main

import "testing"

func TestRunRequiresCandidateInputForCheckDoc(t *testing.T) {
	opt := options{Repo: ".", Manifest: defaultManifest, CheckDoc: true}
	if err := requireCandidateInput(opt); err == nil {
		t.Fatal("expected missing candidate input to fail")
	}
}
