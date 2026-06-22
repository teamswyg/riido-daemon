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
