package supervisor

import (
	"context"
	"sync"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

type Actor struct {
	cfg Config

	mailbox   chan envelope
	stopReqCh chan lifecycle.ShutdownLevel
	stoppedCh chan struct{}
	stopErrCh chan error

	claimMu     sync.Mutex
	claimCancel context.CancelFunc
}
