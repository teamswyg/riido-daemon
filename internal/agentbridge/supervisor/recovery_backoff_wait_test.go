package supervisor

import (
	"context"
	"testing"
	"testing/synctest"
	"time"
)

func TestRecoveryWaitCancellationPreventsAnotherAttempt(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		done := make(chan bool, 1)
		go func() { done <- recoveryWait(ctx, time.Hour) }()
		// Enter the timer wait before cancelling, without depending on scheduling.
		synctest.Wait()
		select {
		case <-done:
			t.Fatal("recovery wait finished before cancellation")
		default:
		}
		cancel()
		synctest.Wait()
		select {
		case elapsed := <-done:
			if elapsed {
				t.Fatal("cancelled recovery wait authorized another attempt")
			}
		default:
			t.Fatal("recovery wait ignored cancellation")
		}
	})
}
