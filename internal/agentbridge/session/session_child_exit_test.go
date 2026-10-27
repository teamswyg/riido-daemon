package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestFinishWaitsForChildExitAfterProviderResult(t *testing.T) {
	p := newChildExitRunning()
	r := newChildExitRunner(context.Background(), p, time.Second)
	go r.finish()
	awaitChildExitSignal(t, p.killStarted)
	select {
	case <-r.sess.Result():
		t.Fatal("provider result and successful Kill do not confirm child exit")
	case <-time.After(20 * time.Millisecond):
	}
	emitChildExit(p, 137)
	result := waitResult(t, r.sess, time.Second)
	if result.Status != agentbridge.ResultCompleted || result.Output != "provider result" {
		t.Fatalf("child cleanup changed provider result: %+v", result)
	}
	if !r.sess.childExitObserved {
		t.Fatal("actual child exit was not observed")
	}
}

func TestFinishMissingChildExitRemainsUnknown(t *testing.T) {
	for _, status := range []agentbridge.ResultStatus{agentbridge.ResultCompleted, agentbridge.ResultFailed, agentbridge.ResultCancelled, agentbridge.ResultTimeout} {
		t.Run(string(status), func(t *testing.T) {
			p := newChildExitRunning()
			r := newChildExitRunner(context.Background(), p, 10*time.Millisecond)
			r.state.Result.Status = status
			go r.finish()
			if result := waitResult(t, r.sess, time.Second); result.Status != status {
				t.Fatalf("cleanup changed result: %s", result.Status)
			}
			if r.sess.childExitObserved {
				t.Fatal("terminal result alone became exit evidence")
			}
			emitChildExit(p, 0)
			if r.sess.childExitObserved {
				t.Fatal("late exit revised the published observation")
			}
		})
	}
}

func TestFinishClosedChannelsWithoutStatusRemainUnknown(t *testing.T) {
	p := newChildExitRunning()
	close(p.stdout)
	close(p.stderr)
	close(p.exited)
	r := newChildExitRunner(context.Background(), p, time.Second)
	go r.finish()
	waitResult(t, r.sess, time.Second)
	if r.sess.childExitObserved {
		t.Fatal("channel closure became exit evidence")
	}
}

func TestFinishPreservesAlreadyObservedChildExit(t *testing.T) {
	p := newChildExitRunning()
	close(p.stdout)
	close(p.stderr)
	r := newChildExitRunner(context.Background(), p, time.Second)
	r.deferExit(process.ExitStatus{Code: 23, Err: errors.New("child failed")}, true)
	go r.finish()
	waitResult(t, r.sess, time.Second)
	if !r.sess.childExitObserved {
		t.Fatal("previous core exit observation was lost")
	}
}

func TestFinishDrainsBothStreamsWhileKillWaits(t *testing.T) {
	p := newChildExitRunning()
	ioDone := make(chan struct{})
	p.killFn = func(ctx context.Context) error {
		select {
		case <-ioDone:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	r := newChildExitRunner(context.Background(), p, time.Second)
	go r.finish()
	awaitChildExitSignal(t, p.killStarted)
	producerDone := make(chan struct{})
	go func() {
		defer close(producerDone)
		defer close(ioDone)
		timer := time.NewTimer(time.Second)
		defer timer.Stop()
		for range 128 {
			select {
			case p.stdout <- []byte("late stdout"):
			case <-timer.C:
				return
			}
			select {
			case p.stderr <- []byte("late stderr"):
			case <-timer.C:
				return
			}
		}
		emitChildExit(p, 0)
	}()
	waitResult(t, r.sess, 2*time.Second)
	awaitChildExitSignal(t, producerDone)
	if !r.sess.childExitObserved {
		t.Fatal("stream drain did not unblock child exit")
	}
	parser := r.parser.(*recordingParser)
	if len(parser.stdoutChunks) != 0 || len(parser.stderrChunks) != 0 || len(r.sess.events) != 0 {
		t.Fatal("cleanup parsed or published late provider output")
	}
}

func TestFinishBoundsDrainWhenExitChannelClosesWithoutStatus(t *testing.T) {
	p := newChildExitRunning()
	close(p.exited)
	r := newChildExitRunner(context.Background(), p, 10*time.Millisecond)
	stop := make(chan struct{})
	producerDone := make(chan struct{})
	go func() {
		defer close(producerDone)
		for {
			select {
			case p.stdout <- []byte("continuous stdout"):
			case p.stderr <- []byte("continuous stderr"):
			case <-stop:
				return
			}
		}
	}()
	go r.finish()
	waitResult(t, r.sess, time.Second)
	close(stop)
	awaitChildExitSignal(t, producerDone)
	if r.sess.childExitObserved {
		t.Fatal("a closed exit channel and active output became evidence")
	}
}

func TestFinishKillErrorDoesNotDetermineChildExit(t *testing.T) {
	for _, actualExit := range []bool{false, true} {
		p := newChildExitRunning()
		p.killFn = func(context.Context) error {
			if actualExit {
				emitChildExit(p, 23)
			}
			return errors.New("kill failed")
		}
		r := newChildExitRunner(context.Background(), p, 10*time.Millisecond)
		go r.finish()
		waitResult(t, r.sess, time.Second)
		if r.sess.childExitObserved != actualExit {
			t.Fatalf("Kill error determined evidence; actual exit=%t", actualExit)
		}
	}
}

func TestFinishUsesOneKillAndObservationBudget(t *testing.T) {
	p := newChildExitRunning()
	killReturned := make(chan struct{})
	p.killFn = func(ctx context.Context) error { <-ctx.Done(); close(killReturned); return ctx.Err() }
	r := newChildExitRunner(context.Background(), p, 200*time.Millisecond)
	go r.finish()
	awaitChildExitSignal(t, killReturned)
	// Kill consumed the whole budget; observation must not start another one.
	waitResult(t, r.sess, 100*time.Millisecond)
	if r.sess.childExitObserved {
		t.Fatal("kill timeout became exit evidence")
	}
}

func TestFinishObservesExitDespiteCancelledRunContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := newChildExitRunning()
	killHadLiveContext := false
	p.killFn = func(ctx context.Context) error {
		killHadLiveContext = ctx.Err() == nil
		emitChildExit(p, 137)
		return nil
	}
	r := newChildExitRunner(ctx, p, time.Second)
	go r.finish()
	waitResult(t, r.sess, time.Second)
	if !killHadLiveContext || !r.sess.childExitObserved {
		t.Fatal("run cancellation prevented bounded child-exit observation")
	}
}

func TestChildExitObservationStaysWithItsSession(t *testing.T) {
	p1, p2 := newChildExitRunning(), newChildExitRunning()
	r1 := newChildExitRunner(context.Background(), p1, time.Second)
	r2 := newChildExitRunner(context.Background(), p2, 10*time.Millisecond)
	emitChildExit(p1, 0)
	go r1.finish()
	go r2.finish()
	waitResult(t, r1.sess, time.Second)
	waitResult(t, r2.sess, time.Second)
	if !r1.sess.childExitObserved || r2.sess.childExitObserved {
		t.Fatal("child-exit observation crossed Session boundaries")
	}
}

func TestChildExitObservationIsReadyBeforeSessionDone(t *testing.T) {
	scenario := startToolGateScenario(t, "assignment-child-exit", noEventRecordingAdapter(), nil)
	scenario.session.Cancel(nil)
	awaitChildExitSignal(t, scenario.session.Done())
	if !scenario.session.childExitObserved {
		t.Fatal("fake Kill emitted an exit but Session did not observe it")
	}
	if result := waitResult(t, scenario.session, time.Second); result.Status != agentbridge.ResultCancelled {
		t.Fatalf("cleanup changed cancellation result: %s", result.Status)
	}
}

func TestStartFailureDoesNotProduceChildExitObservation(t *testing.T) {
	sess, err := Start(context.Background(), Config{Adapter: noEventRecordingAdapter(), Process: childExitStartFailure{}})
	if err == nil || sess != nil {
		t.Fatal("process-start failure must not produce a Session or exit evidence")
	}
}
