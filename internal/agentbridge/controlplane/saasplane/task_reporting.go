package saasplane

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// accumulatePartialBody appends a text delta to the task's evolving assistant
// body and, on a debounce boundary (time or character growth), forwards the
// full text-so-far as one tagged progress line. Raw per-delta fragments are
// never forwarded on their own.
func (p *Plane) accumulatePartialBody(ctx context.Context, assignment assignmentcontract.Assignment, executionID, delta string) error {
	if delta == "" {
		return nil
	}
	var (
		flush bool
		body  string
	)
	now := time.Now()
	if err := p.withState(ctx, func(s *planeState) {
		st := s.partialBodies[executionID]
		if st == nil {
			st = &partialBodyState{}
			s.partialBodies[executionID] = st
		}
		st.text += delta
		grown := len(st.text) - st.lastFlushedLen
		if st.lastFlushAt.IsZero() || now.Sub(st.lastFlushAt) >= partialBodyFlushInterval || grown >= partialBodyFlushChars {
			flush = true
			st.lastFlushAt = now
			st.lastFlushedLen = len(st.text)
			body = st.text
		}
	}); err != nil {
		return err
	}
	if !flush {
		return nil
	}
	return p.postPartialBody(ctx, assignment, body)
}

// postPartialBody forwards the accumulated body as a single evolving progress
// line, reusing the existing EventProgress → riido_log mapping. The sentinel
// progress code keeps the body verbatim and the progress key reaches the client
// as the line's message_key — no control-plane change required.
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

func (p *Plane) CompleteTask(ctx context.Context, executionID string, res agentbridge.Result) error {
	assignment, ok, err := p.assignmentForExecution(ctx, executionID)
	if err != nil || !ok {
		return err
	}
	state, eventType := terminalStateAndEvent(res.Status)
	message := res.Error
	if message == "" {
		message = res.Output
	}
	// Some providers (e.g. codex completing via thread/status/changed) report a
	// successful result with no Output. Fall back to the accumulated streamed
	// body so the completed thread shows the actual answer instead of an empty
	// message (which the client renders as a generic status label).
	if message == "" && res.Status == agentbridge.ResultCompleted {
		if stateErr := p.withState(ctx, func(s *planeState) {
			if st := s.partialBodies[executionID]; st != nil {
				message = st.text
			}
		}); stateErr != nil {
			return stateErr
		}
	}
	_, err = p.postAgentEvent(ctx, assignment, assignmentcontract.AgentEventRequest{
		AssignmentID:      assignment.ID,
		TaskID:            assignment.TaskID,
		ProviderSessionID: res.SessionID,
		State:             state,
		EventType:         eventType,
		Message:           message,
	})
	if err != nil {
		return err
	}
	return p.withState(ctx, func(s *planeState) {
		closeCancelWatcher(s, executionID)
		delete(s.assignmentsByExecution, executionID)
		delete(s.runtimeIDsByExecution, executionID)
		delete(s.partialBodies, executionID)
	})
}

func (p *Plane) pollAgent(ctx context.Context, agentID, runtimeID string, wait time.Duration) (assignmentcontract.PollResponse, error) {
	var out assignmentcontract.PollResponse
	err := p.postJSON(ctx, "/v1/agents/"+url.PathEscape(agentID)+"/poll", assignmentcontract.PollRequest{
		DaemonID:  p.cfg.DaemonID,
		DeviceID:  p.cfg.DeviceID,
		RuntimeID: runtimeID,
		WaitMs:    pollWaitMilliseconds(wait),
	}, &out)
	return out, err
}

func pollWaitMilliseconds(wait time.Duration) int {
	if wait <= 0 {
		return 0
	}
	milliseconds := wait.Milliseconds()
	if milliseconds <= 0 {
		return 1
	}
	return int(milliseconds)
}

func (p *Plane) postAgentEvent(ctx context.Context, assignment assignmentcontract.Assignment, req assignmentcontract.AgentEventRequest) (assignmentcontract.AgentEventResponse, error) {
	var out assignmentcontract.AgentEventResponse
	req.DaemonID = p.cfg.DaemonID
	req.DeviceID = p.cfg.DeviceID
	runtimeID, err := p.runtimeIDForAssignment(ctx, assignment)
	if err != nil {
		return out, err
	}
	req.RuntimeID = runtimeID
	err = p.postJSON(ctx, "/v1/agents/"+url.PathEscape(assignment.AgentID)+"/events", req, &out)
	return out, err
}

func (p *Plane) postJSON(ctx context.Context, path string, in, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return p.doJSON(ctx, http.MethodPost, path, body, out)
}

func (p *Plane) getJSON(ctx context.Context, path string, out any) error {
	return p.doJSON(ctx, http.MethodGet, path, nil, out)
}
