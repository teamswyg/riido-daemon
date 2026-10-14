package saasplane

import (
	"strings"

	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
)

func normalizeAgents(in []AgentBinding) []AgentBinding {
	out := make([]AgentBinding, 0, len(in))
	for _, agent := range in {
		agent.AgentID = strings.TrimSpace(agent.AgentID)
		agent.RuntimeProvider = providercatalog.String(agent.RuntimeProvider)
		if agent.AgentID == "" || agent.RuntimeProvider == "" {
			continue
		}
		out = append(out, agent)
	}
	return out
}

func (p *Plane) dynamicBindingsEnabled() bool {
	return len(p.cfg.Agents) == 0
}
