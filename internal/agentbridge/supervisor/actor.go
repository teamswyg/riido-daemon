package supervisor

import (
	"context"
	"sync"

	"github.com/teamswyg/riido-daemon/internal/execution"
	"github.com/teamswyg/riido-daemon/internal/providerrecovery"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

type Actor struct {
	cfg      Config
	recovery *providerrecovery.Circuits

	mailbox   chan envelope
	stopReqCh chan lifecycle.ShutdownLevel
	stoppedCh chan struct{}
	stopErrCh chan error

	claimMu                 sync.Mutex
	claimCancel             context.CancelFunc
	registeredRuntimeStates map[string]string
	heartbeatMu             sync.RWMutex
	heartbeatExecutions     map[string][]string
	executionObservations   []execution.Observation
}
