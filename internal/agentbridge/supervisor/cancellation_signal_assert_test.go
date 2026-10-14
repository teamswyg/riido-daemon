package supervisor

import (
	"testing"
	"time"
)

func expectSignal(t *testing.T, ch <-chan struct{}, msg string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(supervisorCancellationTestTimeout):
		t.Fatal(msg)
	}
}
