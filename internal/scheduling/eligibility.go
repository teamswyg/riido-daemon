package scheduling

import (
	"fmt"
	"sort"
	"strings"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

// RequiredSurface is a provider-neutral task requirement. The names are
// owned by docs/20-domain/runtime-scheduling.md §1.
type RequiredSurface string

const (
	SurfaceStructuredEventStream RequiredSurface = "structured-event-stream"
	SurfaceSessionResume         RequiredSurface = "session-resume"
	SurfaceSystemPrompt          RequiredSurface = "system-prompt"
	SurfaceMaxTurns              RequiredSurface = "max-turns"
	SurfaceMCP                   RequiredSurface = "mcp"
	SurfaceToolHooks             RequiredSurface = "tool-hooks"
	SurfaceUsage                 RequiredSurface = "usage"
	SurfaceWorktree              RequiredSurface = "worktree"
)

// TaskRequirements is the C5 view of a task before claim/execute.
type TaskRequirements struct {
	Provider                 capability.ProviderKind
	RequiredSurfaces         []RequiredSurface
	AllowExperimentalRuntime bool
}

// RuntimeCapability is the C5 subset of C3 ProviderCapability needed for
// scheduler eligibility. It intentionally excludes process/runtime details.
type RuntimeCapability struct {
	RuntimeID                 capability.RuntimeID
	Provider                  capability.ProviderKind
	CapabilityFingerprint     capability.CapabilityFingerprint
	SlotLimit                 int
	SlotsInUse                int
	Available                 bool
	CompatibilityStatus       capability.CompatibilityStatus
	RequiresExperimentalOptIn bool
	SupportsStreaming         bool
	SupportsResume            bool
	SupportsSystem            bool
	SupportsMaxTurns          bool
	SupportsMCP               bool
	SupportsToolHooks         bool
	SupportsUsage             bool
	SupportsWorktree          bool
}

// IneligibilityReason is one reason a runtime cannot execute a task.
type IneligibilityReason struct {
	Code    string
	Surface RequiredSurface
	Detail  string
}

// Eligibility is the deterministic result of comparing one task's
// requirements against one runtime capability snapshot.
type Eligibility struct {
	Eligible              bool
	RuntimeID             capability.RuntimeID
	CapabilityFingerprint capability.CapabilityFingerprint
	Reasons               []IneligibilityReason
}

// EvaluateCapability applies the C5 pre-submit scheduling gate.
func EvaluateCapability(req TaskRequirements, cap RuntimeCapability) Eligibility {
	out := Eligibility{
		Eligible:              true,
		RuntimeID:             cap.RuntimeID,
		CapabilityFingerprint: cap.CapabilityFingerprint,
	}
	if req.Provider != "" && req.Provider != cap.Provider {
		out.add("PROVIDER_MISMATCH", "", fmt.Sprintf("task provider %q does not match runtime provider %q", req.Provider, cap.Provider))
	}
	if !cap.Available {
		out.add("PROVIDER_UNAVAILABLE", "", fmt.Sprintf("provider %q is unavailable", cap.Provider))
	}
	if cap.CompatibilityStatus == capability.CompatBlocked {
		out.add("COMPATIBILITY_BLOCKED", "", fmt.Sprintf("provider %q compatibility is blocked", cap.Provider))
	}
	if cap.RequiresExperimentalOptIn && !req.AllowExperimentalRuntime {
		out.add("EXPERIMENTAL_RUNTIME_REQUIRES_OPT_IN", "", fmt.Sprintf("provider %q requires allow_experimental_runtime", cap.Provider))
	}
	if cap.SlotLimit > 0 && cap.SlotsInUse >= cap.SlotLimit {
		out.add("SLOT_EXHAUSTED", "", fmt.Sprintf("runtime %q has no free execution slots", cap.RuntimeID))
	}
	for _, surface := range NormalizeRequiredSurfaces(req.RequiredSurfaces) {
		supported, known := supportsSurface(cap, surface)
		if !known {
			out.add("UNKNOWN_REQUIRED_SURFACE", surface, fmt.Sprintf("unknown required surface %q", surface))
			continue
		}
		if !supported {
			out.add("MISSING_REQUIRED_SURFACE", surface, fmt.Sprintf("provider %q does not support required surface %q", cap.Provider, surface))
		}
	}
	return out
}

func (e *Eligibility) add(code string, surface RequiredSurface, detail string) {
	e.Eligible = false
	e.Reasons = append(e.Reasons, IneligibilityReason{Code: code, Surface: surface, Detail: detail})
}

// Summary returns a stable human-readable reason string for logs/results.
func (e Eligibility) Summary() string {
	if e.Eligible {
		return "eligible"
	}
	parts := make([]string, 0, len(e.Reasons))
	for _, reason := range e.Reasons {
		if reason.Surface != "" {
			parts = append(parts, fmt.Sprintf("%s:%s", reason.Code, reason.Surface))
			continue
		}
		parts = append(parts, reason.Code)
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

// NormalizeRequiredSurfaces makes task-supplied surface names
// deterministic and de-duplicated. Unknown values are preserved so the
// evaluator can fail closed.
func NormalizeRequiredSurfaces(in []RequiredSurface) []RequiredSurface {
	seen := map[RequiredSurface]bool{}
	out := make([]RequiredSurface, 0, len(in))
	for _, surface := range in {
		normalized := RequiredSurface(strings.ReplaceAll(strings.ToLower(strings.TrimSpace(string(surface))), "_", "-"))
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func supportsSurface(cap RuntimeCapability, surface RequiredSurface) (bool, bool) {
	switch surface {
	case SurfaceStructuredEventStream:
		return cap.SupportsStreaming, true
	case SurfaceSessionResume:
		return cap.SupportsResume, true
	case SurfaceSystemPrompt:
		return cap.SupportsSystem, true
	case SurfaceMaxTurns:
		return cap.SupportsMaxTurns, true
	case SurfaceMCP:
		return cap.SupportsMCP, true
	case SurfaceToolHooks:
		return cap.SupportsToolHooks, true
	case SurfaceUsage:
		return cap.SupportsUsage, true
	case SurfaceWorktree:
		return cap.SupportsWorktree, true
	default:
		return false, false
	}
}
