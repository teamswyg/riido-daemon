package process

import (
	"context"
	"errors"
)

func (r *FakeRunning) WriteStdin(b []byte) error {
	select {
	case r.stdin <- append([]byte(nil), b...):
		return nil
	default:
		return errors.New("fake stdin buffer full")
	}
}

func (r *FakeRunning) CloseStdin() error {
	defer func() { _ = recover() }() // tolerate double-close
	close(r.stdin)
	return nil
}

func (r *FakeRunning) Kill(_ context.Context) error {
	select {
	case r.kill <- struct{}{}:
	default:
	}
	r.emitToActor(fakeEmit{kind: fakeEmitExit, status: ExitStatus{Code: 137, Err: errors.New("killed")}})
	return nil
}
