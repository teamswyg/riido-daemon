package saasplane

import (
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestPlaneUsesSharedAssignmentContractSurface(t *testing.T) {
	if assignmentcontract.SchemaVersion != "riido-ai-server.v1" {
		t.Fatalf("schema version = %q", assignmentcontract.SchemaVersion)
	}
	if !assignmentcontract.PollStart.Valid() || !assignmentcontract.AssignmentReady.Valid() {
		t.Fatal("shared assignment contract validation is not wired")
	}
}
