package supervisor

import (
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

func applyPolicyDefaults(cfg *Config) error {
	if cfg.PolicyBundleVersion == "" {
		cfg.PolicyBundleVersion = cfg.PolicyBundle.Version
		if cfg.PolicyBundleVersion == "" {
			cfg.PolicyBundleVersion = policy.DefaultLocalPolicyBundleVersion
		}
	}
	if cfg.PolicyBundle.SchemaVersion == "" {
		cfg.PolicyBundle = policy.DefaultLocalPolicyBundle()
		cfg.PolicyBundle.Version = cfg.PolicyBundleVersion
		return nil
	}
	if err := cfg.PolicyBundle.Validate(); err != nil {
		return fmt.Errorf("supervisor: policy bundle: %w", err)
	}
	if cfg.PolicyBundleVersion != cfg.PolicyBundle.Version {
		return fmt.Errorf("supervisor: PolicyBundleVersion %q does not match policy bundle version %q", cfg.PolicyBundleVersion, cfg.PolicyBundle.Version)
	}
	return nil
}
