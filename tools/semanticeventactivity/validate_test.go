package main

import "testing"

func TestValidateRejectsUnknownEvent(t *testing.T) {
	manifest := validManifest("doc.md")
	manifest.SemanticActivity = append(manifest.SemanticActivity, "made_up")
	if len(validate(testRepo(t), manifest)) == 0 {
		t.Fatalf("expected unknown event failure")
	}
}

func TestValidateRejectsCategoryDrift(t *testing.T) {
	manifest := validManifest("doc.md")
	manifest.SemanticActivity = remove(manifest.SemanticActivity, "progress")
	manifest.NonSemanticActivity = append(manifest.NonSemanticActivity, "progress")
	if len(validate(testRepo(t), manifest)) == 0 {
		t.Fatalf("expected category drift failure")
	}
}
