package main

import "testing"

func TestCommandMentionsToken(t *testing.T) {
	if !commandMentionsToken("RIIDO_AI_AGENT_TOKEN=x") {
		t.Fatal("token text not detected")
	}
	if commandMentionsToken("-product-storage-state .riido-local/private/state.json") {
		t.Fatal("storage-state should not be treated as token text")
	}
}

func TestSafeCommandPreviewRedactsTokenText(t *testing.T) {
	got := safeCommandPreview("RIIDO_AI_AGENT_TOKEN=x go run ./tools/localqarunner")
	if got != "[redacted: command contains token text]" {
		t.Fatalf("preview=%q", got)
	}
}
