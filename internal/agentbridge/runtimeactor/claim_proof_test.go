package runtimeactor

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type claimBlockedStart struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once
	running *process.FakeRunning
}

func (p *claimBlockedStart) Start(context.Context, process.Command) (process.RunningProcess, error) {
	close(p.entered)
	<-p.release
	return p.running, nil
}
func (p *claimBlockedStart) unblock() { p.once.Do(func() { close(p.release) }) }

func TestClaimProofSubmitReplyLost(t *testing.T) {
	p := &claimBlockedStart{entered: make(chan struct{}), release: make(chan struct{}), running: process.NewFakeRunning()}
	a, err := New(Config{RuntimeID: "claim-lost-reply", Adapters: []agentbridge.Adapter{&stubAdapter{name: "fake", detected: agentbridge.DetectResult{Available: true}}}, Process: p})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		p.unblock()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = a.Stop(ctx)
	})
	scope := executionproof.NewScope(true, nil)
	runCtx := executionproof.ContextWithScope(context.Background(), scope)
	admission, cancel := context.WithCancel(context.Background())
	defer cancel()
	returned := make(chan error, 1)
	go func() {
		handle, submitErr := a.SubmitPreflight(admission, runCtx, bridge.TaskRequest{ID: "claim", Provider: "fake"})
		if handle != nil {
			returned <- errors.New("unexpected handle")
			return
		}
		returned <- submitErr
	}()
	select {
	case <-p.entered:
	case <-time.After(time.Second):
		t.Fatal("Start did not enter")
	}
	cancel()
	select {
	case err = <-returned:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("Submit did not return")
	}
	// The caller lost its handle while Process.Start was still running. A tracker
	// populated from returned handles alone would incorrectly prove zero starts.
	if _, ok := scope.Tracker().Seal(); ok {
		t.Fatal("lost Submit handle omitted an admitted invocation")
	}
	p.unblock()
	stopCtx, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if err = a.Stop(stopCtx); err != nil {
		t.Fatal(err)
	}
	if _, ok := scope.Tracker().Seal(); ok {
		t.Fatal("late actual exit upgraded sealed unknown")
	}
}
