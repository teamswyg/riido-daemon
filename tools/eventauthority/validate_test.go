package main

import "testing"

func TestValidateRejectsAssignedFieldInDraft(t *testing.T) {
	repo := testRepo(t)
	manifest := validManifest("doc.md")
	manifest.IngestorAssignedFields = append(manifest.IngestorAssignedFields, "TaskID")
	problems, _, _ := validate(repo, manifest)
	if len(problems) == 0 {
		t.Fatalf("expected assigned field exposure failure")
	}
}

func TestValidateRejectsMissingBuilderField(t *testing.T) {
	repo := testRepo(t)
	manifest := validManifest("doc.md")
	manifest.DraftSuppliedFields = append(manifest.DraftSuppliedFields, "RunID")
	problems, _, _ := validate(repo, manifest)
	if len(problems) == 0 {
		t.Fatalf("expected missing builder field failure")
	}
}
