package supervisor

import (
	"errors"
	"time"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

func normalizeConfig(cfg Config) (Config, error) {
	if err := validateRequiredConfig(cfg); err != nil {
		return cfg, err
	}
	applyTimingDefaults(&cfg)
	if cfg.RiidoDaemonVersion == "" {
		cfg.RiidoDaemonVersion = "riido-agentd v0.0.0"
	}
	if err := applyPolicyDefaults(&cfg); err != nil {
		return cfg, err
	}
	if cfg.RuntimeTrustTier == "" {
		cfg.RuntimeTrustTier = policy.TrustTierHost
	}
	return cfg, nil
}

func validateRequiredConfig(cfg Config) error {
	if cfg.DaemonID == "" {
		return errors.New("supervisor: DaemonID is required")
	}
	if len(configuredRuntimes(cfg)) == 0 {
		return errors.New("supervisor: at least one Runtime is required")
	}
	if cfg.Source == nil {
		return errors.New("supervisor: Source is required")
	}
	if cfg.Reporter == nil {
		return errors.New("supervisor: Reporter is required")
	}
	return nil
}

func applyTimingDefaults(cfg *Config) {
	if cfg.PollEvery <= 0 {
		cfg.PollEvery = time.Second
	}
	if cfg.IdlePollEvery <= 0 {
		cfg.IdlePollEvery = cfg.PollEvery
	}
	if cfg.IdlePollEvery < cfg.PollEvery {
		cfg.IdlePollEvery = cfg.PollEvery
	}
	if cfg.HeartbeatEvery <= 0 {
		cfg.HeartbeatEvery = 5 * time.Second
	}
	if cfg.MailboxSize <= 0 {
		cfg.MailboxSize = DefaultMailboxSize
	}
}
