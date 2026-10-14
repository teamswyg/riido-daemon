package ingest

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
)

func assertRedactedEvent(t *testing.T, redacted ir.CanonicalEvent, eventID string) {
	t.Helper()
	if redacted.EventID != eventID {
		t.Fatalf("returned event must be redacted event: %s vs %+v", eventID, redacted)
	}
	if strings.Contains(redacted.Payload["text"].(string), "ghp_") {
		t.Fatalf("payload leaked raw github token: %+v", redacted.Payload)
	}
	if strings.Contains(redacted.Unknown["raw"].(string), "RIIDO_TOKEN=") {
		t.Fatalf("unknown leaked raw env token: %+v", redacted.Unknown)
	}
	nested := redacted.Payload["nested"].(map[string]any)
	if strings.Contains(nested["url"].(string), "user:pass") {
		t.Fatalf("nested payload leaked basic auth URL: %+v", nested)
	}
}
