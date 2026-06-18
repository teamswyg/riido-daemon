package saasplane

import (
	"context"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func (p *Plane) failUnresumableActiveAssignment(ctx context.Context, assignment assignmentcontract.Assignment) error {
	_, err := p.postAgentEvent(ctx, assignment, assignmentcontract.AgentEventRequest{
		AssignmentID: assignment.ID,
		TaskID:       assignment.TaskID,
		State:        assignmentcontract.AssignmentFailed,
		EventType:    assignmentcontract.EventAssignmentFailed,
		Message:      "recovery requires provider session id; refusing fresh start",
		Metadata: map[string]string{
			metadatakeys.AssignmentRecovery.String(): assignmentcontract.RecoveryFreshStartRefused.String(),
		},
	})
	return err
}
