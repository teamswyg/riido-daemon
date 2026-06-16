package runtimeactor

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	providercap "github.com/teamswyg/riido-contracts/provider/capability"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

type capabilityProfile struct {
	protocolKind              providercap.ProtocolKind
	protocolMaturity          providercap.ProtocolMaturity
	eventStreamFormat         providercap.EventStreamFormat
	supportsPermissionControl bool
	exposesUnsafeBypass       bool
	supportsApprovalProtocol  bool
	supportsFileEvents        bool
	supportsWorktree          bool
	defaultSandboxMode        string
	defaultApprovalPolicy     string
}

func buildRuntimeCapability(runtimeID, provider string, res agentbridge.DetectResult, policyBundleVersion string, discoveredAt time.Time) (Capability, error) {
	domain, err := reconcileProviderCapability(runtimeID, provider, res, policyBundleVersion, discoveredAt)
	if err != nil {
		return Capability{}, err
	}
	return Capability{
		Provider:                  provider,
		Available:                 res.Available,
		Version:                   res.Version,
		Executable:                res.Executable,
		Profile:                   metaProfile(res.Metadata),
		Reason:                    res.Reason,
		ProtocolKind:              string(domain.ProtocolKind),
		AdapterID:                 domain.AdapterID,
		AdapterVersion:            domain.AdapterVersion,
		ProtocolVersion:           domain.ProtocolVersion,
		CompatibilityStatus:       string(domain.CompatibilityStatus),
		CapabilityFingerprint:     string(domain.CapabilityFingerprint),
		DetectedFingerprint:       string(domain.DetectedFingerprint),
		RequiresExperimentalOptIn: domain.RequiresExperimentalOptIn,
		SupportsStreaming:         domain.SupportsStructuredEventStream,
		SupportsResume:            domain.SupportsResume,
		SupportsSystem:            domain.SupportsSystemPrompt,
		SupportsMaxTurns:          domain.SupportsMaxTurns,
		SupportsMCP:               domain.SupportsMCP,
		SupportsToolHooks:         domain.SupportsHookEvents,
		SupportsUsage:             domain.SupportsUsageMetrics,
		SupportsFileEvents:        domain.SupportsFileEvents,
		SupportsWorktree:          domain.SupportsWorktree,
	}, nil
}

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

	domain := providercap.ProviderCapability{
		RuntimeID:                     providercap.RuntimeID(runtimeID),
		ProviderKind:                  providercap.ProviderKind(provider),
		ProtocolKind:                  profile.protocolKind,
		AdapterID:                     provider,
		AdapterVersion:                "riido-agentbridge-adapter.v1",
		ProtocolVersion:               "v1",
		ExecutablePath:                res.Executable,
		Argv0:                         res.Executable,
		DetectedVersion:               res.Version,
		DetectedFingerprint:           detectedFingerprintForExecutable(res.Executable),
		DiscoveredAt:                  discoveredAt,
		SupportsStructuredEventStream: res.SupportsStreaming,
		EventStreamFormat:             profile.eventStreamFormat,
		SupportsPartialDeltas:         res.SupportsStreaming,
		SupportsResume:                res.SupportsResume,
		SupportsSessionID:             res.SupportsResume,
		SupportsSessionPin:            providercatalog.IsCodex(provider),
		SupportsSystemPrompt:          res.SupportsSystem,
		SupportsMaxTurns:              res.SupportsMaxTurns,
		SupportsToolEvents:            res.SupportsToolHooks,
		SupportsUsageMetrics:          res.SupportsUsage,
		SupportsFileEvents:            profile.supportsFileEvents,
		SupportsPermissionControl:     profile.supportsPermissionControl,
		ExposesUnsafePermissionBypass: profile.exposesUnsafeBypass,
		SupportsApprovalProtocol:      profile.supportsApprovalProtocol,
		SupportsMCP:                   res.SupportsMCP,
		SupportsHookEvents:            res.SupportsToolHooks,
		SupportsWorktree:              profile.supportsWorktree,
		DefaultSandboxMode:            profile.defaultSandboxMode,
		DefaultApprovalPolicy:         profile.defaultApprovalPolicy,
		CompatibilityStatus:           status,
		ProtocolMaturity:              profile.protocolMaturity,
		RequiresExperimentalOptIn:     requiresExperimental,
		MissingCapabilities:           missing,
		BlockedReasons:                blocked,
		DegradedReasons:               degraded,
		Unknown: map[string]any{
			"detect_metadata": copyMetadata(res.Metadata),
		},
	}

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

func detectedFingerprintForExecutable(executable string) providercap.DetectedFingerprint {
	executable = strings.TrimSpace(executable)
	if executable == "" || !filepath.IsAbs(executable) {
		return ""
	}
	info, err := os.Stat(executable)
	if err != nil || !info.Mode().IsRegular() {
		return ""
	}
	file, err := os.Open(executable)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}
	return providercap.DetectedFingerprint(hex.EncodeToString(hash.Sum(nil)))
}
