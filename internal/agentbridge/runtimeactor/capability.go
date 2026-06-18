package runtimeactor

import (
	"time"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func reconcileProviderCapability(runtimeID, provider string, res agentbridge.DetectResult, policyBundleVersion string, discoveredAt time.Time) (providercap.ProviderCapability, error) {
	profile := profileForProvider(provider)
	missing := missingCapabilities(res)
	blocked := []providercap.CompatibilityReason{}
	degraded := []providercap.CompatibilityReason{}
	if !res.Available {
		blocked = append(blocked, providercap.CompatibilityReason{
			Code:    "DETECTION_UNAVAILABLE",
			Subject: provider,
			Detail:  textutil.FirstNonEmpty(res.Reason, "provider detect reported unavailable"),
		})
	} else if len(missing) > 0 {
		degraded = append(degraded, providercap.CompatibilityReason{
			Code:    "SURFACE_PARTIAL",
			Subject: provider,
			Detail:  "one or more provider-neutral capability flags are false",
		})
	}

	status := summarizeCompatibility(profile.protocolMaturity, blocked, degraded)
	requiresExperimental := profile.protocolMaturity == providercap.ProtocolMaturityExperimental ||
		profile.protocolMaturity == providercap.ProtocolMaturityDeprecated

	domain := newProviderCapability(providerCapabilityInput{
		runtimeID:            runtimeID,
		provider:             provider,
		res:                  res,
		profile:              profile,
		status:               status,
		requiresExperimental: requiresExperimental,
		missingCapabilities:  missing,
		blockedReasons:       blocked,
		degradedReasons:      degraded,
		discoveredAt:         discoveredAt,
		policyBundleVersion:  policyBundleVersion,
	})

	fp, err := providercap.ComputeCapabilityFingerprint(providercap.CapabilityFingerprintInput{
		ProviderKind:          domain.ProviderKind,
		ProtocolKind:          domain.ProtocolKind,
		ProviderVersion:       domain.DetectedVersion,
		DetectedFingerprint:   domain.DetectedFingerprint,
		AdapterID:             domain.AdapterID,
		AdapterVersion:        domain.AdapterVersion,
		ProtocolVersion:       domain.ProtocolVersion,
		DefaultSandboxMode:    domain.DefaultSandboxMode,
		DefaultApprovalPolicy: domain.DefaultApprovalPolicy,
		PolicyBundleVersion:   policyBundleVersion,
		ImportantSurfaceFlags: importantSurfaceFlags(domain),
	})
	if err != nil {
		return providercap.ProviderCapability{}, err
	}
	domain.CapabilityFingerprint = fp
	return domain, nil
}
