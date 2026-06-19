package main

import "github.com/teamswyg/riido-daemon/internal/policy"

func validatePolicy(rows []Surface) ([]problem, []PolicyEvidence) {
	known := knownPolicySurfaces()
	var problems []problem
	evidence := make([]PolicyEvidence, 0, len(rows))
	for _, row := range rows {
		surface := policy.UnsafeBypassSurface(row.Surface)
		ok := known[surface] && hostUnknownDeny(surface) && isolatedRequiresBundle(surface)
		evidence = append(evidence, PolicyEvidence{Surface: row.Surface, OK: ok})
		if !ok {
			problems = append(problems, problem{Message: "unsafe bypass policy drift for " + row.Surface})
		}
	}
	return problems, evidence
}

func knownPolicySurfaces() map[policy.UnsafeBypassSurface]bool {
	return map[policy.UnsafeBypassSurface]bool{
		policy.UnsafeBypassClaudePermissions: true,
		policy.UnsafeBypassCursorYolo:        true,
		policy.UnsafeBypassCodexYolo:         true,
		policy.UnsafeBypassCodexDangerBypass: true,
	}
}
