package supervisor

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type policyBundleScenario struct {
	running *process.FakeRunning
	result  agentbridge.Result
}

type policyBundleScenarioConfig struct {
	provider                 string
	taskID                   string
	bundleVersion            string
	allowExperimentalRuntime bool
}

func noSurfacePolicyBundle(version string) policy.PolicyBundle {
	return policy.PolicyBundle{
		SchemaVersion:  policy.BundleSchemaVersion,
		Version:        version,
		EffectiveSince: time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		TrustTierPolicies: map[policy.TrustTier]policy.TrustTierPolicy{
			policy.TrustTierHost: {AllowedSurfaces: policy.AllowedSurfaceSet{}},
		},
	}
}
