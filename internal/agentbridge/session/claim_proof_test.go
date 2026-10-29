package session

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestClaimProofObservesStartAndExitBeforeResult(t *testing.T) {
	scope := executionproof.NewScope(true, nil)
	p := process.NewFake()
	sess, err := Start(executionproof.ContextWithScope(context.Background(), scope), Config{Adapter: noEventRecordingAdapter(), Process: p})
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for range sess.Events() {
		}
	}()
	sess.Cancel(nil)
	waitResult(t, sess, time.Second)
	proof, ok := scope.Tracker().Seal()
	if !ok || proof.StartedInvocations != 1 {
		t.Fatalf("actual Session start/exit was not accumulated: %+v, %t", proof, ok)
	}
}

func TestClaimProofRejectsStartAfterSeal(t *testing.T) {
	scope := executionproof.NewScope(true, nil)
	scope.Tracker().Seal()
	p := process.NewFake()
	sess, err := Start(executionproof.ContextWithScope(context.Background(), scope), Config{Adapter: noEventRecordingAdapter(), Process: p})
	if sess != nil {
		go func() {
			for range sess.Events() {
			}
		}()
		sess.Cancel(nil)
		waitResult(t, sess, time.Second)
	}
	if err == nil {
		t.Fatal("sealed claim admitted another Process.Start")
	}
}

func TestClaimProofStartErrorUnknown(t *testing.T) {
	scope := executionproof.NewScope(true, nil)
	_, err := Start(executionproof.ContextWithScope(context.Background(), scope), Config{Adapter: noEventRecordingAdapter(), Process: childExitStartFailure{}})
	if err == nil {
		t.Fatal("expected generic start failure")
	}
	if _, ok := scope.Tracker().Seal(); ok {
		t.Fatal("generic Process.Start error became zero-start proof")
	}
}

type claimFixedProcess struct{ running process.RunningProcess }

func (p claimFixedProcess) Start(context.Context, process.Command) (process.RunningProcess, error) {
	return p.running, nil
}

func TestClaimProofContinuationPreservesUnknown(t *testing.T) {
	scope := executionproof.NewScope(true, nil)
	ctx := executionproof.ContextWithScope(context.Background(), scope)
	first, err := Start(ctx, Config{Adapter: noEventRecordingAdapter(), Process: claimFixedProcess{running: newChildExitRunning()}, ProcessKillTimeout: 10 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for range first.Events() {
		}
	}()
	first.Cancel(nil)
	waitResult(t, first, time.Second)
	second, err := Start(ctx, Config{Adapter: noEventRecordingAdapter(), Process: process.NewFake()})
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for range second.Events() {
		}
	}()
	second.Cancel(nil)
	waitResult(t, second, time.Second)
	if !second.childExitObserved {
		t.Fatal("second Session fixture did not observe exit")
	}
	if _, ok := scope.Tracker().Seal(); ok {
		t.Fatal("last Session exit replaced earlier unknown observation")
	}
}
