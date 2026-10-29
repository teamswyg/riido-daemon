package saasplane

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

type planeState struct {
	executionScopes         map[string]*executionproof.Scope
	assignmentsByExecution  map[string]assignmentcontract.Assignment
	runtimeIDsByExecution   map[string]string
	cancelWatchers          map[string]chan error
	registeredRuntimes      map[string]RuntimeSnapshotRecord
	registeredDeviceName    string
	lastRuntimeSnapshotSync time.Time
	agentBindingsCache      []assignmentcontract.AgentRuntimeBinding
	agentBindingsCachedAt   time.Time
	connectionRevision      string
	connectedPrincipalCount int
	nextAssignmentEventSeq  uint64
	// partialBodies accumulates each execution's assistant text deltas between
	// flushes so the daemon can forward a coherent evolving body instead of
	// per-token fragments. Keyed by execution ID.
	partialBodies map[string]*partialBodyState
}

// partialBodyState holds the running assistant text and process-local delivery
// timings for one execution.
type partialBodyState struct {
	text             string
	claimedAt        time.Time
	firstTextAt      time.Time
	firstDeliveredAt time.Time
}

type stateOp struct {
	fn    func(*planeState)
	close bool
	ack   chan struct{}
}

func newPlaneState() planeState {
	return planeState{
		executionScopes:        map[string]*executionproof.Scope{},
		assignmentsByExecution: map[string]assignmentcontract.Assignment{},
		runtimeIDsByExecution:  map[string]string{},
		cancelWatchers:         map[string]chan error{},
		registeredRuntimes:     map[string]RuntimeSnapshotRecord{},
		partialBodies:          map[string]*partialBodyState{},
	}
}
