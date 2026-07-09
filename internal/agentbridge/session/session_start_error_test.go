package session

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestStartReturnsProcessStartError(t *testing.T) {
	want := errors.New("spawn denied")
	_, err := Start(context.Background(), Config{
		Adapter: &recordingAdapter{},
		Process: failingStartProcess{err: want},
	})
	if err == nil {
		t.Fatal("expected start error")
	}
	if !errors.Is(err, want) {
		t.Fatalf("wrapped error=%v, want %v", err, want)
	}
	if !strings.Contains(err.Error(), "session: process start") {
		t.Fatalf("missing stable prefix in %q", err)
	}
}

type failingStartProcess struct {
	err error
}

func (p failingStartProcess) Start(context.Context, process.Command) (process.RunningProcess, error) {
	return nil, p.err
}
