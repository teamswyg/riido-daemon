package main

import "testing"

func TestDecodeObjectPayloadWrapsArrayResponses(t *testing.T) {
	got := decodeObjectPayload([]byte(`[{"id":"team-a"}]`))
	if firstArrayObjectString(got["items"], "id") != "team-a" {
		t.Fatalf("decoded=%v", got)
	}
}
