package process

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

// EmitStdout queues a stdout chunk for consumers of Stdout().
func (r *FakeRunning) EmitStdout(b []byte) {
	r.emitToActor(fakeEmit{kind: fakeEmitStdout, bytes: append([]byte(nil), b...)})
}

// EmitStderr queues a stderr chunk.
func (r *FakeRunning) EmitStderr(b []byte) {
	r.emitToActor(fakeEmit{kind: fakeEmitStderr, bytes: append([]byte(nil), b...)})
}

// EmitExit signals process termination and closes stdout/stderr/exited
// channels. Subsequent Emit* calls are ignored; the fake owns channel closure
// in one actor goroutine so burst tests can race Emit* and EmitExit.
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
