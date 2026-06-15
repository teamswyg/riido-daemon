package scheduling

import (
	"fmt"
	"slices"
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
func EvaluateCapability(req TaskRequirements, candidate RuntimeCapability) Eligibility {
	out := Eligibility{
		Eligible:              true,
		RuntimeID:             candidate.RuntimeID,
		CapabilityFingerprint: candidate.CapabilityFingerprint,
	}
	if req.Provider != "" && req.Provider != candidate.Provider {
		out.add("PROVIDER_MISMATCH", "", fmt.Sprintf("task provider %q does not match runtime provider %q", req.Provider, candidate.Provider))
	}
	if !candidate.Available {
		out.add("PROVIDER_UNAVAILABLE", "", fmt.Sprintf("provider %q is unavailable", candidate.Provider))
	}
	if candidate.CompatibilityStatus == capability.CompatBlocked {
		out.add("COMPATIBILITY_BLOCKED", "", fmt.Sprintf("provider %q compatibility is blocked", candidate.Provider))
	}
	if candidate.RequiresExperimentalOptIn && !req.AllowExperimentalRuntime {
		out.add("EXPERIMENTAL_RUNTIME_REQUIRES_OPT_IN", "", fmt.Sprintf("provider %q requires allow_experimental_runtime", candidate.Provider))
	}
	if candidate.SlotLimit > 0 && candidate.SlotsInUse >= candidate.SlotLimit {
		out.add("SLOT_EXHAUSTED", "", fmt.Sprintf("runtime %q has no free execution slots", candidate.RuntimeID))
	}
	for _, surface := range NormalizeRequiredSurfaces(req.RequiredSurfaces) {
		supported, known := supportsSurface(candidate, surface)
		if !known {
			out.add("UNKNOWN_REQUIRED_SURFACE", surface, fmt.Sprintf("unknown required surface %q", surface))
			continue
		}
		if !supported {
			out.add("MISSING_REQUIRED_SURFACE", surface, fmt.Sprintf("provider %q does not support required surface %q", candidate.Provider, surface))
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
	slices.Sort(out)
	return out
}

func supportsSurface(candidate RuntimeCapability, surface RequiredSurface) (bool, bool) {
	switch surface {
	case SurfaceStructuredEventStream:
		return candidate.SupportsStreaming, true
	case SurfaceSessionResume:
		return candidate.SupportsResume, true
	case SurfaceSystemPrompt:
		return candidate.SupportsSystem, true
	case SurfaceMaxTurns:
		return candidate.SupportsMaxTurns, true
	case SurfaceMCP:
		return candidate.SupportsMCP, true
	case SurfaceToolHooks:
		return candidate.SupportsToolHooks, true
	case SurfaceUsage:
		return candidate.SupportsUsage, true
	case SurfaceWorktree:
		return candidate.SupportsWorktree, true
	default:
		return false, false
	}
}
