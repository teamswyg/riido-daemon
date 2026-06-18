package ingest

import (
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
)

func assertRedactionAuditEvent(t *testing.T, audit ir.CanonicalEvent, sourceEventID string) {
	t.Helper()
	if audit.Type != ir.EventPolicyViolationDetected {
		t.Fatalf("second event must be audit event: %+v", audit)
	}
	if audit.ActorKind != ir.ActorAgent || audit.ActorID != "run-1" {
		t.Fatalf("audit attribution mismatch: %+v", audit)
	}
	assertRedactionAuditPayload(t, audit, sourceEventID)
}

func assertRedactionAuditPayload(t *testing.T, audit ir.CanonicalEvent, sourceEventID string) {
	t.Helper()
	if audit.Payload["category"] != "SECRET_LEAK_ATTEMPTED" || audit.Payload["severity"] != "high" {
		t.Fatalf("audit payload mismatch: %+v", audit.Payload)
	}
	if audit.Payload["sourceEventID"] != sourceEventID || audit.Payload["sourceEventType"] != string(ir.EventTextDelta) {
		t.Fatalf("audit source mismatch: %+v", audit.Payload)
	}
	assertRedactionAuditSubject(t, audit)
	assertRedactionAuditFields(t, audit)
}

func assertRedactionAuditSubject(t *testing.T, audit ir.CanonicalEvent) {
	t.Helper()
	subject, _ := audit.Payload["subject"].(string)
	for _, want := range []string{"basic-auth-url", "env-secret-assignment", "github-token"} {
		if !strings.Contains(subject, want) {
			t.Fatalf("audit subject %q missing %q", subject, want)
		}
	}
}

func assertRedactionAuditFields(t *testing.T, audit ir.CanonicalEvent) {
	t.Helper()
	redactedFields, ok := audit.Payload["redactedFields"].([]string)
	if !ok {
		t.Fatalf("redactedFields type = %T", audit.Payload["redactedFields"])
	}
	for _, want := range []string{"payload.text", "payload.nested.url", "unknown.raw"} {
		if !slices.Contains(redactedFields, want) {
			t.Fatalf("redacted fields %v missing %q", redactedFields, want)
		}
	}
}
