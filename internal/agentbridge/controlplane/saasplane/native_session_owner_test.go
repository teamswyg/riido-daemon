package saasplane

import (
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestAssignmentCarriesNativeSessionOwnership(t *testing.T) {
	assignment := assignmentcontract.Assignment{ID: "assignment", WorkspaceID: "workspace", TaskID: "task", AgentID: "agent-owner", RuntimeProvider: "claude"}
	req := taskRequestFromAssignment(assignment)
	if req.Metadata[controlplane.MetadataAgentID] != assignment.AgentID || req.Metadata[MetadataAgentID] != assignment.AgentID {
		t.Fatalf("native session owner/report identity missing: %+v", req.Metadata)
	}
}
