package controlplane

import (
	"sync"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

// MemorySource is the simplest TaskSourcePort: tasks live in a FIFO,
// runtimes in a map. Intended for tests, offline mode, and bootstrap.
//
// Adapter-local state is guarded because liveness, claim and cancellation
// callers may operate concurrently. No lock is held across network/process IO.
type MemorySource struct {
	mu        sync.Mutex
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
