package saasplane

import (
	"context"
	"strings"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// accumulatePartialBody appends a text delta to the task's evolving assistant
// body before attempting any network operation, then forwards the snapshot.
func (p *Plane) accumulatePartialBody(ctx context.Context, assignment assignmentcontract.Assignment, executionID, delta string) error {
	if delta == "" {
		return nil
	}
	var body string
	accepted := false
	if err := p.withState(ctx, func(s *planeState) {
		if !currentClaimScope(ctx, s, executionID) {
			return
		}
		accepted = true
		st := s.partialBodies[executionID]
		if st == nil {
			st = &partialBodyState{}
			s.partialBodies[executionID] = st
		}
		st.text += delta
		if st.firstTextAt.IsZero() && strings.TrimSpace(st.text) != "" {
			st.firstTextAt = time.Now()
			logResponseTiming(executionID, "first_text", st.claimedAt, st.firstTextAt)
		}
		body = st.text
	}); err != nil {
		return err
	}
	if !accepted {
		return nil
	}
	// A confirmation transition can fail too. Keep the text so the next
	// snapshot or terminal result still contains this first fragment.
	if assignment.ConfirmationReason != "" {
		if _, err := p.postAgentEvent(ctx, assignment, assignmentcontract.AgentEventRequest{AssignmentID: assignment.ID, TaskID: assignment.TaskID, State: assignmentcontract.AssignmentRunning, EventType: assignmentcontract.EventAssignmentRunning}); err != nil {
			return err
		}
	}
	if err := p.postPartialBody(ctx, assignment, body); err != nil {
		return err
	}
	return p.withState(ctx, func(s *planeState) {
		if !currentClaimScope(ctx, s, executionID) {
			return
		}
		st := s.partialBodies[executionID]
		if st != nil && st.firstDeliveredAt.IsZero() && strings.TrimSpace(body) != "" {
			st.firstDeliveredAt = time.Now()
			logResponseTiming(executionID, "first_delivery", st.claimedAt, st.firstDeliveredAt)
		}
	})
}

// postPartialBody forwards the accumulated body as one evolving progress line.
func (p *Plane) postPartialBody(ctx context.Context, assignment assignmentcontract.Assignment, body string) error {
	req, ok := eventRequestFromAgentEvent(assignment, agentbridge.Event{
		Kind:         agentbridge.EventProgress,
		Text:         body,
		ProgressCode: assistantPartialProgressCode,
		ProgressKey:  assistantPartialProgressKey,
	})
	if !ok {
		return nil
	}
	_, err := p.postAgentEvent(ctx, assignment, req)
	return err
}
