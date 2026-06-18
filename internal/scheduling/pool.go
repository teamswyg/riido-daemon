package scheduling

import "sort"

// RuntimeSelection is the deterministic result of selecting one runtime
// from a pool for a task.
type RuntimeSelection struct {
	Runtime     RuntimeCapability
	Eligibility Eligibility
	Rejected    []Eligibility
}

// SelectRuntime evaluates a runtime pool and returns the best eligible
// candidate. It is pure C5 logic: no locks, leases, persistence, or
// provider execution side effects happen here.
func SelectRuntime(req TaskRequirements, candidates []RuntimeCapability) (RuntimeSelection, bool) {
	eligible, rejected := classifyCandidates(req, candidates)
	sort.Slice(rejected, func(i, j int) bool {
		return eligibilityLess(rejected[i], rejected[j])
	})
	if len(eligible) == 0 {
		return RuntimeSelection{Rejected: rejected}, false
	}
	sort.Slice(eligible, func(i, j int) bool {
		return runtimeCandidateLess(eligible[i].Runtime, eligible[j].Runtime)
	})
	selected := eligible[0]
	selected.Rejected = rejected
	return selected, true
}

func classifyCandidates(req TaskRequirements, candidates []RuntimeCapability) ([]RuntimeSelection, []Eligibility) {
	eligible := []RuntimeSelection{}
	rejected := []Eligibility{}
	for _, candidate := range candidates {
		evaluation := EvaluateCapability(req, candidate)
		if evaluation.Eligible {
			eligible = append(eligible, RuntimeSelection{Runtime: candidate, Eligibility: evaluation})
			continue
		}
		rejected = append(rejected, evaluation)
	}
	return eligible, rejected
}
