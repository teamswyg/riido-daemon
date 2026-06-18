package scheduling

import (
	"slices"
	"strings"
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

// NormalizeRequiredSurfaces makes task-supplied surface names
// deterministic and de-duplicated. Unknown values are preserved so the
// evaluator can fail closed.
func NormalizeRequiredSurfaces(in []RequiredSurface) []RequiredSurface {
	seen := map[RequiredSurface]bool{}
	out := make([]RequiredSurface, 0, len(in))
	for _, surface := range in {
		normalized := normalizeSurface(surface)
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	slices.Sort(out)
	return out
}

func normalizeSurface(surface RequiredSurface) RequiredSurface {
	value := strings.ToLower(strings.TrimSpace(string(surface)))
	return RequiredSurface(strings.ReplaceAll(value, "_", "-"))
}
