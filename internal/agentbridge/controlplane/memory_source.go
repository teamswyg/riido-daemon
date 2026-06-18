package controlplane

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

// MemorySource is the simplest TaskSourcePort: tasks live in a FIFO,
// runtimes in a map. Intended for tests, offline mode, and bootstrap.
//
// All state is owned by the calling goroutine -- the source itself is
// NOT a separate actor. Callers (daemon main goroutine or a
// SupervisorActor) serialize access. We do not use sync.Mutex here.
type MemorySource struct {
	queue     []bridge.TaskRequest
	runtimes  map[string]*RegisteredRuntime
	cancelChs map[string]chan error
	now       func() time.Time
}

func NewMemorySource() *MemorySource {
	return &MemorySource{
		runtimes:  map[string]*RegisteredRuntime{},
		cancelChs: map[string]chan error{},
		now:       time.Now,
	}
}
