package saasplane

import (
	"context"
	"net/url"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

const runtimeIDAgentMarker = "agent"

type agentRuntimeID struct {
	DaemonID string
	AgentID  string
	Provider providercatalog.Kind
}

func newAgentRuntimeID(daemonID string, agent AgentBinding) agentRuntimeID {
	return agentRuntimeID{
		DaemonID: strings.TrimSpace(daemonID),
		AgentID:  strings.TrimSpace(agent.AgentID),
		Provider: providercatalog.Normalize(agent.RuntimeProvider),
	}
}

func (id agentRuntimeID) String() string {
	return id.DaemonID + ":" + runtimeIDAgentMarker + ":" + url.QueryEscape(id.AgentID) + ":" + string(id.Provider)
}

func RuntimeIDForAgent(daemonID string, agent AgentBinding) string {
	return newAgentRuntimeID(daemonID, agent).String()
}

func parseAgentRuntimeID(runtimeID string) (agentRuntimeID, bool) {
	parts := strings.Split(runtimeID, ":")
	if len(parts) < 4 || parts[len(parts)-3] != runtimeIDAgentMarker {
		return agentRuntimeID{}, false
	}
	agentID, err := url.QueryUnescape(strings.TrimSpace(parts[len(parts)-2]))
	if err != nil {
		return agentRuntimeID{}, false
	}
	id := agentRuntimeID{
		DaemonID: strings.TrimSpace(strings.Join(parts[:len(parts)-3], ":")),
		AgentID:  strings.TrimSpace(agentID),
		Provider: providercatalog.Normalize(parts[len(parts)-1]),
	}
	return id, id.AgentID != ""
}

func providerFromRuntimeID(runtimeID string) string {
	if id, ok := parseAgentRuntimeID(runtimeID); ok {
		return string(id.Provider)
	}
	parts := strings.Split(runtimeID, ":")
	return providercatalog.String(parts[len(parts)-1])
}

func (p *Plane) runtimeIDForAssignment(ctx context.Context, assignment assignmentcontract.Assignment) (string, error) {
	if p.dynamicBindingsEnabled() {
		var runtimeID string
		err := p.withState(ctx, func(s *planeState) {
			runtimeID = s.runtimeIDsByExecution[assignmentExecutionID(assignment)]
		})
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(runtimeID) != "" {
			return runtimeID, nil
		}
	}
	for _, agent := range p.cfg.Agents {
		if agent.AgentID == assignment.AgentID &&
			providercatalog.Normalize(agent.RuntimeProvider) == providercatalog.Normalize(assignment.RuntimeProvider) {
			return RuntimeIDForAgent(p.cfg.DaemonID, agent), nil
		}
	}
	return RuntimeIDForAgent(p.cfg.DaemonID, AgentBinding{AgentID: assignment.AgentID, RuntimeProvider: assignment.RuntimeProvider}), nil
}

func agentFromRuntimeID(runtimeID string) (string, bool) {
	id, ok := parseAgentRuntimeID(runtimeID)
	return id.AgentID, ok
}

func assignmentExecutionID(assignment assignmentcontract.Assignment) string {
	return textutil.FirstNonEmptyTrimmed(assignment.ID, assignment.TaskID)
}
