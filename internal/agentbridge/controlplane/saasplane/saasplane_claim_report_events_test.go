package saasplane

import (
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func assertClaimReportEvents(t *testing.T, fake *fakeAssignmentServer) {
	t.Helper()
	fake.assertEvent(t, assignmentcontract.EventAssignmentReady)
	fake.assertEvent(t, assignmentcontract.EventRiidoLog)
	fake.assertEvent(t, assignmentcontract.EventAssignmentRunning)
	fake.assertEvent(t, assignmentcontract.EventProviderSessionPinned)
	fake.assertEvent(t, assignmentcontract.EventAssignmentCompleted)

	var sawPinned, sawCompleted bool
	for _, event := range fake.events {
		if event.EventType == assignmentcontract.EventProviderSessionPinned && event.ProviderSessionID == "sess-1" {
			sawPinned = true
		}
		if event.EventType == assignmentcontract.EventAssignmentCompleted && event.ProviderSessionID == "sess-1" {
			sawCompleted = true
		}
	}
	if !sawPinned || !sawCompleted {
		t.Fatalf("provider session id was not carried through events: %+v", fake.events)
	}
}
