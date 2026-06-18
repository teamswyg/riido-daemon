package runtimeactor

import (
	"errors"
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

// New validates Config and returns an Actor that has not yet started.
func New(cfg Config) (*Actor, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	cfg = defaultConfig(cfg)
	return &Actor{
		cfg:       cfg,
		mailbox:   make(chan envelope, cfg.MailboxSize),
		statusCh:  make(chan statusMsg, 4),
		stopReqCh: make(chan lifecycle.ShutdownLevel, cfg.MailboxSize),
		stoppedCh: make(chan struct{}),
		stopErrCh: make(chan error, 1),
		startedCh: make(chan struct{}),
	}, nil
}

func validateConfig(cfg Config) error {
	if cfg.RuntimeID == "" {
		return errors.New("runtimeactor: RuntimeID is required")
	}
	if len(cfg.Adapters) == 0 {
		return errors.New("runtimeactor: at least one Adapter is required")
	}
	if cfg.Process == nil {
		return errors.New("runtimeactor: Process port is required")
	}
	return validateAdapterNames(cfg)
}

func validateAdapterNames(cfg Config) error {
	seen := map[string]bool{}
	for _, a := range cfg.Adapters {
		if a.Name() == "" {
			return errors.New("runtimeactor: adapter Name() is empty")
		}
		if seen[a.Name()] {
			return fmt.Errorf("runtimeactor: duplicate adapter name %q", a.Name())
		}
		seen[a.Name()] = true
	}
	return nil
}

func defaultConfig(cfg Config) Config {
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
	return cfg
}
