package process

import (
	"context"
	"errors"
)

// Fake is a deterministic Process used by tests. It spawns no real OS
// process; the test drives stdout / stderr / exit signals through the
// FakeRunning's Emit* methods.
//
// The fake intentionally uses channels for all state to serve as a reference
// for the actor-style discipline used by daemon runtime packages.
type Fake struct {
	// NextRunning lets a test pre-construct the FakeRunning so it can
	// drive stdout/stderr/exit before Start is even called. If nil,
	// Start allocates a fresh FakeRunning.
	NextRunning *FakeRunning
}

func NewFake() *Fake { return &Fake{} }

func (f *Fake) Start(_ context.Context, cmd Command) (RunningProcess, error) {
	r := f.NextRunning
	if r == nil {
		r = NewFakeRunning()
	}
	f.NextRunning = nil
	r.cmd = cmd
	select {
	case r.started <- cmd:
	default:
	}
	return r, nil
}

// FakeRunning is the RunningProcess returned by Fake. Tests inspect
// Command() to verify the adapter built the right spawn command, drive
// EmitStdout / EmitStderr / EmitExit to simulate provider output, and
// read StdinRecv() / KillRecv() to verify what the session actor sent.
type FakeRunning struct {
	cmd     Command
	started chan Command
	stdout  chan []byte
	stderr  chan []byte
	exited  chan ExitStatus
	stdin   chan []byte
	kill    chan struct{}
	emit    chan fakeEmit
	done    chan struct{}
}

func NewFakeRunning() *FakeRunning {
	r := &FakeRunning{
		started: make(chan Command, 1),
		stdout:  make(chan []byte, DefaultStdoutBuffer),
		stderr:  make(chan []byte, DefaultStderrBuffer),
		exited:  make(chan ExitStatus, 1),
		stdin:   make(chan []byte, 64),
		kill:    make(chan struct{}, 1),
		emit:    make(chan fakeEmit, 4096),
		done:    make(chan struct{}),
	}
	go r.runEmitActor()
	return r
}

type fakeEmit struct {
	kind   fakeEmitKind
	bytes  []byte
	status ExitStatus
}

type fakeEmitKind int

const (
	fakeEmitStdout fakeEmitKind = iota + 1
	fakeEmitStderr
	fakeEmitExit
)

func (r *FakeRunning) Stdout() <-chan []byte     { return r.stdout }
func (r *FakeRunning) Stderr() <-chan []byte     { return r.stderr }
func (r *FakeRunning) Exited() <-chan ExitStatus { return r.exited }

// Command returns the spawn command that Start was invoked with. Tests
// use this to assert on the adapter's BuildStart output.
func (r *FakeRunning) Command() Command { return r.cmd }

// StartedRecv lets tests wait until Start has assigned Command(). The
// receive synchronizes with Fake.Start, avoiding racy polling in actor tests.
func (r *FakeRunning) StartedRecv() <-chan Command { return r.started }

// StdinRecv lets tests observe what the session actor wrote to stdin.
func (r *FakeRunning) StdinRecv() <-chan []byte { return r.stdin }

// KillRecv lets tests observe Kill calls.
func (r *FakeRunning) KillRecv() <-chan struct{} { return r.kill }

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

// EmitStdout queues a stdout chunk for consumers of Stdout().
func (r *FakeRunning) EmitStdout(b []byte) {
	r.emitToActor(fakeEmit{kind: fakeEmitStdout, bytes: append([]byte(nil), b...)})
}

// EmitStderr queues a stderr chunk.
func (r *FakeRunning) EmitStderr(b []byte) {
	r.emitToActor(fakeEmit{kind: fakeEmitStderr, bytes: append([]byte(nil), b...)})
}

// EmitExit signals process termination and closes stdout/stderr/exited
// channels. Subsequent Emit* calls are ignored; the fake owns channel
// closure in one actor goroutine so burst tests can race Emit* and EmitExit
// without racing close against send.
func (r *FakeRunning) EmitExit(code int, err error) {
	r.emitToActor(fakeEmit{kind: fakeEmitExit, status: ExitStatus{Code: code, Err: err}})
}

func (r *FakeRunning) emitToActor(event fakeEmit) {
	select {
	case <-r.done:
		return
	default:
	}
	select {
	case <-r.done:
		return
	case r.emit <- event:
	}
}

func (r *FakeRunning) runEmitActor() {
	for event := range r.emit {
		switch event.kind {
		case fakeEmitStdout:
			select {
			case <-r.done:
				return
			case r.stdout <- event.bytes:
			}
		case fakeEmitStderr:
			select {
			case <-r.done:
				return
			case r.stderr <- event.bytes:
			}
		case fakeEmitExit:
			r.exited <- event.status
			close(r.stdout)
			close(r.stderr)
			close(r.exited)
			close(r.done)
			return
		}
	}
}
