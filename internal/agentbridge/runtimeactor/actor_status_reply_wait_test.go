package runtimeactor

import "testing"

func TestRuntimeActorStatusReplyWaitObservesStop(t *testing.T) {
	a := newStoppedReplyWaitActor("rt-status-stop")
	errCh := make(chan error, 1)
	go func() {
		ctx, cancel := replyWaitContext()
		defer cancel()
		status, err := a.Status(ctx)
		if err != nil {
			errCh <- err
			return
		}
		if status.RuntimeID != "rt-status-stop" || status.Health != "stopped" {
			errCh <- errUnexpectedStoppedStatus()
			return
		}
		errCh <- nil
	}()

	waitForStatusReplyWait(t, a.statusCh, "Status")
	close(a.stoppedCh)
	waitForReplyWaitError(t, errCh, "Status")
}

func TestRuntimeActorHeartbeatReplyWaitObservesStop(t *testing.T) {
	a := newStoppedReplyWaitActor("rt-heartbeat-stop")
	errCh := make(chan error, 1)
	go func() {
		ctx, cancel := replyWaitContext()
		defer cancel()
		hb, err := a.HeartbeatPayload(ctx)
		if err != nil {
			errCh <- err
			return
		}
		if hb.RuntimeID != "rt-heartbeat-stop" {
			errCh <- errUnexpectedStoppedHeartbeat()
			return
		}
		errCh <- nil
	}()

	waitForStatusReplyWait(t, a.statusCh, "HeartbeatPayload")
	close(a.stoppedCh)
	waitForReplyWaitError(t, errCh, "HeartbeatPayload")
}
