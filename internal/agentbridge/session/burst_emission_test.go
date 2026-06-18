package session

type burstStdoutEmitter interface {
	EmitStdout([]byte)
}

type burstStderrEmitter interface {
	EmitStderr([]byte)
}

type burstExitEmitter interface {
	EmitExit(int, error)
}

func emitBurstStdout(running burstStdoutEmitter, chunks int) {
	for range chunks {
		running.EmitStdout([]byte("s"))
	}
	running.EmitStdout([]byte("DONE"))
}

func emitBurstStderr(running burstStderrEmitter, chunks int) {
	for range chunks {
		running.EmitStderr([]byte("e"))
	}
}

func safeEmitStdout(running burstStdoutEmitter, chunk []byte) {
	defer func() { _ = recover() }()
	running.EmitStdout(chunk)
}

func safeEmitExit(running burstExitEmitter, code int, err error) {
	defer func() { _ = recover() }()
	running.EmitExit(code, err)
}
