package saasplane

import (
	"log"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Monotonic, process-local durations. Never compare host and server clocks.
func logResponseTiming(executionID, phase string, claimedAt, observedAt time.Time) {
	if claimedAt.IsZero() || observedAt.Before(claimedAt) {
		return
	}
	log.Printf("event=agent_response_timing execution_id=%q phase=%q claim_elapsed_ms=%d", executionID, phase, observedAt.Sub(claimedAt).Milliseconds())
}

func hasVerifiedResult(res agentbridge.Result) bool {
	return res.ExecutionContext != nil && res.ExecutionContext.VerificationStatus == "passed"
}
