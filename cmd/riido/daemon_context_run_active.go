package main

import (
	"context"
	"slices"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/contextbroker"
)

const (
	contextRunActiveTimeout = time.Second
	// contextReadConnDeadline covers the broker's 15s fetch plus the
	// activity check and the reply write.
	contextReadConnDeadline = 25 * time.Second
)

// runtimeRunActive reports whether any runtime still has the run in flight.
// Each runtime gets its own timeout; a runtime that does not answer makes a
// "not found" result inconclusive rather than proof the run ended.
func runtimeRunActive(runtimes []*runtimeactor.Actor) contextbroker.RunActive {
	return func(runKey string) (bool, bool) {
		certain := true
		for _, rt := range runtimes {
			ctx, cancel := context.WithTimeout(context.Background(), contextRunActiveTimeout)
			beat, err := rt.HeartbeatPayload(ctx)
			cancel()
			if err != nil {
				certain = false
				continue
			}
			if slices.Contains(beat.RunningTaskIDs, runKey) {
				return true, true
			}
		}
		return false, certain
	}
}
