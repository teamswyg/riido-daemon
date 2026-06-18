package runtimeactor

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newStoppedReplyWaitActor(runtimeID string) *Actor {
	return &Actor{
		cfg:       Config{RuntimeID: runtimeID},
		statusCh:  make(chan statusMsg, 1),
		stoppedCh: make(chan struct{}),
	}
}

func waitForStatusReplyWait(t *testing.T, statusCh <-chan statusMsg, label string) {
	t.Helper()
	select {
	case <-statusCh:
	case <-time.After(time.Second):
		t.Fatalf("%s did not enter reply wait", label)
	}
}

func waitForReplyWaitError(t *testing.T, errCh <-chan error, label string) {
	t.Helper()
	if err := <-errCh; err != nil {
		t.Fatalf("%s: %v", label, err)
	}
}

func errUnexpectedStoppedStatus() error {
	return errors.New("unexpected stopped status")
}

func errUnexpectedStoppedHeartbeat() error {
	return errors.New("unexpected stopped heartbeat")
}

func replyWaitContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second)
}
