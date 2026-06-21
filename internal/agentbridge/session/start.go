package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

// Start spawns the session actor and returns its Session handle. The caller
// must drain Events until closed or the actor can block on send.
func Start(ctx context.Context, cfg Config) (*Session, error) {
	if err := validateSessionConfig(cfg); err != nil {
		return nil, err
	}
	cfg = defaultSessionConfig(cfg)
	running, err := cfg.Process.Start(ctx, cfg.Spawn)
	if err != nil {
		return nil, fmt.Errorf("session: process start: %w", err)
	}
	sess := newSessionHandle(cfg, running)
	go run(ctx, cfg, sess, running)
	return sess, nil
}

func validateSessionConfig(cfg Config) error {
	if cfg.Adapter == nil {
		return errors.New("session: Adapter is required")
	}
	if cfg.Process == nil {
		return errors.New("session: Process is required")
	}
	return nil
}

func defaultSessionConfig(cfg Config) Config {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.EventBuffer <= 0 {
		cfg.EventBuffer = DefaultEventBuffer
	}
	if cfg.ResultBuffer <= 0 {
		cfg.ResultBuffer = DefaultResultBuffer
	}
	if cfg.ProcessKillTimeout <= 0 {
		cfg.ProcessKillTimeout = DefaultProcessKillTimeout
	}
	return cfg
}

func newSessionHandle(cfg Config, running process.RunningProcess) *Session {
	return &Session{
		events:   make(chan agentbridge.Event, cfg.EventBuffer),
		result:   make(chan agentbridge.Result, cfg.ResultBuffer),
		cancel:   make(chan cancelRequest, 1),
		terminal: make(chan terminalRequest, 1),
		done:     make(chan struct{}),
		running:  running,
	}
}
