package process

// FakeRunning is the RunningProcess returned by Fake. Tests inspect Command()
// to verify spawn command construction, drive Emit* methods to simulate
// provider output, and read StdinRecv / KillRecv for actor assertions.
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

func (r *FakeRunning) Stdout() <-chan []byte     { return r.stdout }
func (r *FakeRunning) Stderr() <-chan []byte     { return r.stderr }
func (r *FakeRunning) Exited() <-chan ExitStatus { return r.exited }

// Command returns the spawn command that Start was invoked with.
func (r *FakeRunning) Command() Command { return r.cmd }

// StartedRecv lets tests wait until Start has assigned Command().
func (r *FakeRunning) StartedRecv() <-chan Command { return r.started }

// StdinRecv lets tests observe what the session actor wrote to stdin.
func (r *FakeRunning) StdinRecv() <-chan []byte { return r.stdin }

// KillRecv lets tests observe Kill calls.
func (r *FakeRunning) KillRecv() <-chan struct{} { return r.kill }
