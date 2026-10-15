package saasplane

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

type readyReportError struct {
	cause     error
	retryable bool
}

func (e *readyReportError) Error() string        { return "assignment Ready acknowledgement is unavailable" }
func (e *readyReportError) Unwrap() error        { return e.cause }
func (e *readyReportError) RetryableReady() bool { return e.retryable }

func (p *Plane) StartTask(ctx context.Context, executionID string) error {
	assignment, ok, err := p.assignmentForExecution(ctx, executionID)
	if err != nil {
		return &readyReportError{cause: err}
	}
	if !ok || assignment.ID == "" || assignment.LeaseToken == "" {
		return &readyReportError{cause: errors.New("ready assignment authority is unavailable")}
	}
	req := assignmentcontract.AgentEventRequest{
		AssignmentID: assignment.ID, TaskID: assignment.TaskID,
		State: assignmentcontract.AssignmentReady, EventType: assignmentcontract.EventAssignmentReady, Message: "daemon ready",
		Metadata: map[string]string{"assignment_event_key": readyEventKey(assignment, executionID), "lease_token": assignment.LeaseToken},
	}
	response, err := p.postAgentEvent(ctx, assignment, req)
	if err != nil {
		return &readyReportError{cause: err, retryable: retryableReadyFailure(err)}
	}
	if err = validateReadyAcknowledgement(assignment, req, response); err != nil {
		return &readyReportError{cause: err}
	}
	return nil
}

func readyEventKey(a assignmentcontract.Assignment, executionID string) string {
	sum := sha256.Sum256(fmt.Appendf(nil, "%s\x00%s\x00%s", a.ID, executionID, a.LeaseToken))
	return fmt.Sprintf("ready:%x", sum[:])
}

func validateReadyAcknowledgement(a assignmentcontract.Assignment, req assignmentcontract.AgentEventRequest, response assignmentcontract.AgentEventResponse) error {
	event := response.Event
	if response.SchemaVersion != assignmentcontract.SchemaVersion || event.AssignmentID != a.ID || event.TaskID != a.TaskID || event.AgentID != a.AgentID || event.Type != assignmentcontract.EventAssignmentReady || event.State != assignmentcontract.AssignmentReady || event.Seq <= 0 || event.EventID != fmt.Sprintf("%s:%d", a.ID, event.Seq) || event.At.IsZero() || event.Metadata["assignment_event_key"] != req.Metadata["assignment_event_key"] || event.Metadata["lease_token"] != "" || event.WorkspaceID != a.WorkspaceID || event.TeamID != a.TeamID || event.CreatedBy != a.CreatedBy {
		return errors.New("ready acknowledgement event identity is invalid")
	}
	acknowledged := response.Assignment
	if acknowledged == nil || acknowledged.ID != a.ID || acknowledged.TaskID != a.TaskID || acknowledged.AgentID != a.AgentID || acknowledged.LeaseToken != a.LeaseToken || acknowledged.State != assignmentcontract.AssignmentReady || acknowledged.RuntimeProvider != a.RuntimeProvider || acknowledged.ModelID != a.ModelID || acknowledged.ComponentID != a.ComponentID || acknowledged.RequestID != a.RequestID || acknowledged.WorkspaceID != a.WorkspaceID || acknowledged.TeamID != a.TeamID || acknowledged.CreatedBy != a.CreatedBy {
		return errors.New("ready acknowledgement assignment lease is invalid")
	}
	return nil
}

func retryableReadyFailure(err error) bool {
	var status httpStatusError
	if errors.As(err, &status) {
		return retryableHTTPStatus(status.StatusCode)
	}
	var transport net.Error
	return errors.As(err, &transport) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF)
}
