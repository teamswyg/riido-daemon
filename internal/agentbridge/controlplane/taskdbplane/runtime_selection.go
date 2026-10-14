package taskdbplane

import (
	"sort"

	"github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/scheduling"
)

func (p *Plane) runtimeSelectionForTask(provider, runtimeID string) (scheduling.RuntimeSelection, bool) {
	if len(p.runtimes) == 0 {
		return scheduling.RuntimeSelection{Runtime: scheduling.RuntimeCapability{
			RuntimeID: capability.RuntimeID(runtimeID),
			Provider:  capability.ProviderKind(provider),
			Available: true,
		}}, true
	}
	selection, ok := scheduling.SelectRuntime(scheduling.TaskRequirements{
		Provider: capability.ProviderKind(provider),
	}, p.runtimeCandidatesForProvider(provider))
	if !ok {
		return scheduling.RuntimeSelection{}, false
	}
	if string(selection.Runtime.RuntimeID) != runtimeID {
		return scheduling.RuntimeSelection{}, false
	}
	return selection, true
}

func (p *Plane) runtimeCandidatesForProvider(provider string) []scheduling.RuntimeCapability {
	ids := make([]string, 0, len(p.runtimes))
	for id := range p.runtimes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]scheduling.RuntimeCapability, 0, len(ids))
	for _, id := range ids {
		candidate, ok := runtimeCapabilityForProvider(p.runtimes[id], provider)
		if ok {
			out = append(out, candidate)
		}
	}
	return out
}
