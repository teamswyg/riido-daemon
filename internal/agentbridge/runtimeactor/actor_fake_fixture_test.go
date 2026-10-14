package runtimeactor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func availableFakeAdapters() []agentbridge.Adapter {
	return []agentbridge.Adapter{
		&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}},
	}
}

func startAvailableFakeActor(t *testing.T, cfg Config) (*Actor, *fakeProcess) {
	t.Helper()
	cfg.Adapters = availableFakeAdapters()
	return startActor(t, cfg)
}

func startManualAvailableFakeActor(t *testing.T, cfg Config) *Actor {
	t.Helper()
	cfg.Adapters = availableFakeAdapters()
	a, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	return a
}
