package scheduling

import (
	"fmt"
	"sort"
	"strings"
)

func (e *Eligibility) add(code string, surface RequiredSurface, detail string) {
	e.Eligible = false
	e.Reasons = append(e.Reasons, IneligibilityReason{Code: code, Surface: surface, Detail: detail})
}

func addSurfaceReasons(out *Eligibility, req TaskRequirements, candidate RuntimeCapability) {
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
}

// Summary returns a stable human-readable reason string for logs/results.
func (e Eligibility) Summary() string {
	if e.Eligible {
		return "eligible"
	}
	parts := make([]string, 0, len(e.Reasons))
	for _, reason := range e.Reasons {
		parts = append(parts, reasonSummary(reason))
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func reasonSummary(reason IneligibilityReason) string {
	if reason.Surface != "" {
		return fmt.Sprintf("%s:%s", reason.Code, reason.Surface)
	}
	return reason.Code
}
