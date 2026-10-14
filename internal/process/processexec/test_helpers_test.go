package processexec

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func drainAll(ch <-chan []byte, deadline time.Duration) []byte {
	var out []byte
	timer := time.NewTimer(deadline)
	defer timer.Stop()
	for {
		select {
		case chunk, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, chunk...)
		case <-timer.C:
			return out
		}
	}
}

func requireExit(t *testing.T, proc process.RunningProcess) process.ExitStatus {
	t.Helper()
	select {
	case status := <-proc.Exited():
		return status
	case <-time.After(2 * time.Second):
		t.Fatal("no exit signal")
		return process.ExitStatus{}
	}
}
