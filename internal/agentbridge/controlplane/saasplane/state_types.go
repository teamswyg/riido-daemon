package saasplane

import (
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

type planeState struct {
	assignmentsByExecution  map[string]assignmentcontract.Assignment
	runtimeIDsByExecution   map[string]string
	cancelWatchers          map[string]chan error
	registeredRuntimes      map[string]RuntimeSnapshotRecord
	registeredDeviceName    string
	lastRuntimeSnapshotSync time.Time
	agentBindingsCache      []assignmentcontract.AgentRuntimeBinding
	agentBindingsCachedAt   time.Time
	nextAssignmentEventSeq  uint64
	// partialBodies accumulates each execution's assistant text deltas between
	// flushes so the daemon can forward a coherent evolving body instead of
	// per-token fragments. Keyed by execution ID.
	partialBodies map[string]*partialBodyState
}

// partialBodyState holds the running assistant text for one task and the
// debounce bookkeeping for forwarding it as an evolving progress line.
type partialBodyState struct {
	text           string
	lastFlushAt    time.Time
	lastFlushedLen int
}

type stateOp struct {
	fn    func(*planeState)
	close bool
	ack   chan struct{}
}

func newPlaneState() planeState {
	return planeState{
		assignmentsByExecution: map[string]assignmentcontract.Assignment{},
		runtimeIDsByExecution:  map[string]string{},
		cancelWatchers:         map[string]chan error{},
		registeredRuntimes:     map[string]RuntimeSnapshotRecord{},
		partialBodies:          map[string]*partialBodyState{},
	}
}
