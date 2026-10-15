package saasplane

import (
	"errors"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func validateRecoveryAcknowledgement(assignment assignmentcontract.Assignment, request assignmentcontract.AgentEventRequest, response assignmentcontract.AgentEventResponse) error {
	event := response.Event
	if response.SchemaVersion != assignmentcontract.SchemaVersion || event.AssignmentID != assignment.ID || event.TaskID != assignment.TaskID || event.AgentID != assignment.AgentID || event.Type != assignmentcontract.EventRiidoLog || event.Seq <= 0 || event.At.IsZero() || event.Metadata["assignment_event_key"] != request.Metadata["assignment_event_key"] {
		return errors.New("provider recovery acknowledgement identity is invalid")
	}
	expected := assignmentcontract.RecoveryContextFromMetadata(request.Metadata)
	observed := assignmentcontract.RecoveryContextFromMetadata(event.Metadata)
	if expected == nil || observed == nil || *expected != *observed {
		return errors.New("provider recovery acknowledgement context is invalid")
	}
	return nil
}
