package runtimeactor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

// Actor is the runtime tier actor.
type Actor struct {
	cfg Config

	// Single owning goroutine writes to capabilities/inFlight/...; all
	// public methods send to mailbox channels.
	mailbox  chan envelope
	statusCh chan statusMsg
	// stopReqCh carries the requested shutdown authority level. Stop
	// callers do a non-blocking send so repeated/concurrent stop requests
	// stay idempotent; while draining, a forced request can still escalate
	// a graceful shutdown.
	stopReqCh chan lifecycle.ShutdownLevel
	stoppedCh chan struct{}
	stopErrCh chan error
	startedCh chan struct{} // closed by Start once the actor loop is live
	startedAt time.Time
}

// envelope is the actor's mailbox shape: one message kind per call.
type envelope struct {
	submit *submitMsg
	cancel *cancelMsg
}

type submitMsg struct {
	ctx   context.Context
	req   bridge.TaskRequest
	reply chan submitReply
}

type submitReply struct {
	handle *SessionHandle
	err    error
}

type cancelMsg struct {
	ctx    context.Context
	taskID string
	reason string
	reply  chan error
}

type statusMsg struct {
	ctx   context.Context
	reply chan statusReply
}

type statusReply struct {
	status Status
	hb     Heartbeat
}

// New validates Config and returns an Actor that has not yet started.
// Call Start(ctx) to begin the actor loop and run Detect on each adapter.
func New(cfg Config) (*Actor, error) {
	if cfg.RuntimeID == "" {
		return nil, errors.New("runtimeactor: RuntimeID is required")
	}
	if len(cfg.Adapters) == 0 {
		return nil, errors.New("runtimeactor: at least one Adapter is required")
	}
	if cfg.Process == nil {
		return nil, errors.New("runtimeactor: Process port is required")
	}
	// Detect duplicate adapter names early.
	seen := map[string]bool{}
	for _, a := range cfg.Adapters {
		if a.Name() == "" {
			return nil, errors.New("runtimeactor: adapter Name() is empty")
		}
		if seen[a.Name()] {
			return nil, fmt.Errorf("runtimeactor: duplicate adapter name %q", a.Name())
		}
		seen[a.Name()] = true
	}
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 4
	}
	if cfg.MailboxSize <= 0 {
		cfg.MailboxSize = DefaultMailboxSize
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.PolicyBundleVersion == "" {
		cfg.PolicyBundleVersion = "policy-bundle.local.v0"
	}
	if cfg.CapabilityRefreshEvery == 0 {
		cfg.CapabilityRefreshEvery = DefaultCapabilityRefreshEvery
	}

	a := &Actor{
		cfg:       cfg,
		mailbox:   make(chan envelope, cfg.MailboxSize),
		statusCh:  make(chan statusMsg, 4),
		stopReqCh: make(chan lifecycle.ShutdownLevel, cfg.MailboxSize),
		stoppedCh: make(chan struct{}),
		stopErrCh: make(chan error, 1),
		startedCh: make(chan struct{}),
	}
	return a, nil
}

// Start runs Detect on every adapter (synchronously, on the caller's
// goroutine) and then launches the actor loop. Returns the first
// Detect error; per-provider Available=false reports do NOT abort.
func (a *Actor) Start(ctx context.Context) error {
	caps := make([]Capability, 0, len(a.cfg.Adapters))
	detectedAt := make(map[string]time.Time, len(a.cfg.Adapters))
	discoveredAt := a.cfg.Now()
	for _, adapter := range a.cfg.Adapters {
		res, err := adapter.Detect(ctx, a.cfg.DetectEnv)
		if err != nil {
			return fmt.Errorf("runtimeactor: Detect %s: %w", adapter.Name(), err)
		}
		capView, err := buildRuntimeCapability(a.cfg.RuntimeID, adapter.Name(), res, a.cfg.PolicyBundleVersion, discoveredAt)
		if err != nil {
			return fmt.Errorf("runtimeactor: capability %s: %w", adapter.Name(), err)
		}
		caps = append(caps, capView)
		detectedAt[adapter.Name()] = discoveredAt
	}
	a.startedAt = discoveredAt
	go a.run(caps, detectedAt)
	close(a.startedCh)
	return nil
}

// run is the actor loop. SOLE owner of in-flight map and per-task state.
func (a *Actor) run(caps []Capability, detectedAt map[string]time.Time) {
	adapters := indexAdapters(a.cfg.Adapters)
	inFlight := map[string]*runningTask{}

	completeCh := make(chan string, 32)

	for {
		select {
		case env := <-a.mailbox:
			switch {
			case env.submit != nil:
				h, err := a.handleSubmit(adapters, caps, detectedAt, inFlight, completeCh, env.submit)
				env.submit.reply <- submitReply{handle: h, err: err}
			case env.cancel != nil:
				env.cancel.reply <- a.handleCancel(inFlight, env.cancel)
			}

		case taskID := <-completeCh:
			delete(inFlight, taskID)

		case msg := <-a.statusCh:
			a.refreshDueCapabilities(msg.ctx, adapters, caps, detectedAt)
			msg.reply <- statusReply{
				status: a.buildStatus(caps, inFlight),
				hb:     a.buildHeartbeat(inFlight),
			}

		case level := <-a.stopReqCh:
			a.drainAndShutdown(lifecycle.NormalizeShutdownLevel(level), inFlight, completeCh)
			close(a.stoppedCh)
			return
		}
	}
}

// runningTask is the actor's per-task bookkeeping.
type runningTask struct {
	taskID   string
	provider string
	handle   *SessionHandle
}
