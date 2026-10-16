package supervisor

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestObsoleteCancellationCannotCancelReplacementClaim(t *testing.T) {
	oldCtx, cancelOld := context.WithCancel(context.Background())
	cancelOld()
	newCtx, cancelNew := context.WithCancel(t.Context())
	defer cancelNew()
	replacement := &runningTask{taskID: "same", ctx: newCtx, cancel: cancelNew}
	pending := map[string]*runningTask{"same": replacement}
	actor := &Actor{}
	actor.handleTaskCancel(t.Context(), &cancelMsg{taskID: "same", claimContext: oldCtx, cause: errors.New("old claim cancelled")}, pending)
	if newCtx.Err() != nil || replacement.cancelCause != nil {
		t.Fatal("obsolete cancellation stopped the replacement claim")
	}
}

func TestCurrentClaimCancellationStillCancelsPendingPreparation(t *testing.T) {
	claimCtx, cancelClaim := context.WithCancel(t.Context())
	defer cancelClaim()
	task := &runningTask{taskID: "same", ctx: claimCtx, cancel: cancelClaim}
	actor := &Actor{}
	actor.handleTaskCancel(t.Context(), &cancelMsg{taskID: "same", claimContext: claimCtx, cause: errors.New("user cancelled")}, map[string]*runningTask{"same": task})
	if claimCtx.Err() == nil || task.cancelCause == nil {
		t.Fatal("current cancellation failed to stop preparation")
	}
}

func TestCancellationWatchCarriesOriginalClaimIntoMailbox(t *testing.T) {
	claimCtx, cancelClaim := context.WithCancel(t.Context())
	defer cancelClaim()
	source := &cancelSource{cancel: make(chan error, 1)}
	actor := &Actor{cfg: Config{Source: source}, mailbox: make(chan envelope, 1), stoppedCh: make(chan struct{})}
	done := make(chan struct{})
	go func() { defer close(done); actor.forwardCancellation(claimCtx, "same", cancelClaim) }()
	source.cancel <- errors.New("user cancelled")
	select {
	case message := <-actor.mailbox:
		if message.cancel == nil || message.cancel.claimContext != claimCtx {
			t.Fatal("cancellation watch lost its claim generation")
		}
		if claimCtx.Err() == nil {
			t.Fatal("watch did not promptly cancel its captured claim")
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation watch did not forward")
	}
	<-done
}
