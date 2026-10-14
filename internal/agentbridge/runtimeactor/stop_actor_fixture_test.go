package runtimeactor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func startStoppableActor(t *testing.T) *Actor {
	t.Helper()
	actor, _ := startActor(t, Config{
		Adapters: []agentbridge.Adapter{
			&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
		},
	})
	return actor
}
