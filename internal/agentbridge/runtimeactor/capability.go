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
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type capabilityProfile struct {
	protocolKind              providercap.ProtocolKind
	protocolMaturity          providercap.ProtocolMaturity
	eventStreamFormat         providercap.EventStreamFormat
	supportsPermissionControl bool
	exposesUnsafeBypass       bool
	supportsApprovalProtocol  bool
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
			Detail:  firstNonEmpty(res.Reason, "provider detect reported unavailable"),
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
		SupportsSessionPin:            provider == "codex",
		SupportsSystemPrompt:          res.SupportsSystem,
		SupportsMaxTurns:              res.SupportsMaxTurns,
		SupportsToolEvents:            res.SupportsToolHooks,
		SupportsUsageMetrics:          res.SupportsUsage,
		SupportsPermissionControl:     profile.supportsPermissionControl,
		ExposesUnsafePermissionBypass: profile.exposesUnsafeBypass,
		SupportsApprovalProtocol:      profile.supportsApprovalProtocol,
		SupportsMCP:                   res.SupportsMCP,
		SupportsHookEvents:            res.SupportsToolHooks,
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

func profileForProvider(provider string) capabilityProfile {
	switch provider {
	case "claude":
		return capabilityProfile{
			protocolKind:              providercap.ProtocolClaudeStreamJSON,
			protocolMaturity:          providercap.ProtocolMaturityStable,
			eventStreamFormat:         providercap.EventStreamFormatNDJSON,
			supportsPermissionControl: true,
			exposesUnsafeBypass:       true,
			defaultSandboxMode:        "unknown",
			defaultApprovalPolicy:     "on-request",
		}
	case "codex":
		return capabilityProfile{
			protocolKind:              providercap.ProtocolCodexAppServer,
			protocolMaturity:          providercap.ProtocolMaturityExperimental,
			eventStreamFormat:         providercap.EventStreamFormatJSONRPCNotifications,
			supportsPermissionControl: true,
			exposesUnsafeBypass:       true,
			supportsApprovalProtocol:  true,
			defaultSandboxMode:        "workspace-write",
			defaultApprovalPolicy:     "on-request",
		}
	case "openclaw":
		return capabilityProfile{
			protocolKind:          providercap.ProtocolOpenClawAgentJSON,
			protocolMaturity:      providercap.ProtocolMaturityExperimental,
			eventStreamFormat:     providercap.EventStreamFormatNDJSON,
			defaultSandboxMode:    "unknown",
			defaultApprovalPolicy: "unknown",
		}
	case "cursor":
		return capabilityProfile{
			protocolKind:          providercap.ProtocolCursorAgentStreamJSON,
			protocolMaturity:      providercap.ProtocolMaturityExperimental,
			eventStreamFormat:     providercap.EventStreamFormatNDJSON,
			exposesUnsafeBypass:   true,
			defaultSandboxMode:    "unknown",
			defaultApprovalPolicy: "unknown",
		}
	default:
		return capabilityProfile{
			protocolKind:          providercap.ProtocolKind(provider + "-unknown"),
			protocolMaturity:      providercap.ProtocolMaturityUnknown,
			eventStreamFormat:     providercap.EventStreamFormatUnknown,
			defaultSandboxMode:    "unknown",
			defaultApprovalPolicy: "unknown",
		}
	}
}

func missingCapabilities(res agentbridge.DetectResult) []providercap.CapabilityName {
	checks := []struct {
		name providercap.CapabilityName
		ok   bool
	}{
		{"structured-event-stream", res.SupportsStreaming},
		{"session-resume", res.SupportsResume},
		{"system-prompt", res.SupportsSystem},
		{"max-turns", res.SupportsMaxTurns},
		{"mcp", res.SupportsMCP},
		{"tool-hooks", res.SupportsToolHooks},
		{"usage", res.SupportsUsage},
	}
	out := []providercap.CapabilityName{}
	for _, check := range checks {
		if !check.ok {
			out = append(out, check.name)
		}
	}
	return out
}

func summarizeCompatibility(maturity providercap.ProtocolMaturity, blocked, degraded []providercap.CompatibilityReason) providercap.CompatibilityStatus {
	switch {
	case len(blocked) > 0:
		return providercap.CompatBlocked
	case maturity == providercap.ProtocolMaturityExperimental || maturity == providercap.ProtocolMaturityDeprecated:
		return providercap.CompatExperimental
	case len(degraded) > 0:
		return providercap.CompatDegraded
	default:
		return providercap.CompatSupported
	}
}

func importantSurfaceFlags(c providercap.ProviderCapability) map[string]any {
	return map[string]any{
		"SupportsStructuredEventStream": c.SupportsStructuredEventStream,
		"EventStreamFormat":             c.EventStreamFormat,
		"SupportsPartialDeltas":         c.SupportsPartialDeltas,
		"SupportsResume":                c.SupportsResume,
		"SupportsSessionID":             c.SupportsSessionID,
		"SupportsSessionPin":            c.SupportsSessionPin,
		"SupportsSystemPrompt":          c.SupportsSystemPrompt,
		"SupportsMaxTurns":              c.SupportsMaxTurns,
		"SupportsToolEvents":            c.SupportsToolEvents,
		"SupportsFileEvents":            c.SupportsFileEvents,
		"SupportsUsageMetrics":          c.SupportsUsageMetrics,
		"SupportsPermissionControl":     c.SupportsPermissionControl,
		"ExposesUnsafePermissionBypass": c.ExposesUnsafePermissionBypass,
		"SupportsApprovalProtocol":      c.SupportsApprovalProtocol,
		"SupportsSandbox":               c.SupportsSandbox,
		"SupportsManagedSettings":       c.SupportsManagedSettings,
		"SupportsHookEvents":            c.SupportsHookEvents,
		"SupportsMCP":                   c.SupportsMCP,
		"SupportsWorktree":              c.SupportsWorktree,
		"SupportsJSONSchemaTools":       c.SupportsJSONSchemaTools,
		"ProtocolMaturity":              c.ProtocolMaturity,
		"CompatibilityStatus":           c.CompatibilityStatus,
		"RequiresExperimentalOptIn":     c.RequiresExperimentalOptIn,
	}
}

func copyMetadata(in map[string]string) map[string]string {
	if in == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func firstNonEmpty(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
