package runtimeactor

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestNewRequiresRuntimeID(t *testing.T) {
	_, err := New(Config{
		Adapters: []agentbridge.Adapter{&stubAdapter{name: "x"}},
		Process:  newFakeProcess(),
	})
	if err == nil {
		t.Fatal("expected error without RuntimeID")
	}
}

func TestNewRequiresAtLeastOneAdapter(t *testing.T) {
	_, err := New(Config{RuntimeID: "rt-1", Process: newFakeProcess()})
	if err == nil {
		t.Fatal("expected error without adapters")
	}
}

func TestNewRequiresProcessPort(t *testing.T) {
	_, err := New(Config{
		RuntimeID: "rt-1",
		Adapters:  []agentbridge.Adapter{&stubAdapter{name: "x"}},
	})
	if err == nil {
		t.Fatal("expected error without Process")
	}
}
