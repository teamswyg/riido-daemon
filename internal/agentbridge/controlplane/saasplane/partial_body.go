package saasplane

import (
	"context"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// accumulatePartialBody appends a text delta to the task's evolving assistant
// body and, on a debounce boundary, forwards the full text-so-far as one line.
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
