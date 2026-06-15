package saasplane

import (
	"context"
	"net/url"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func providerModelOverride(runtimeProvider, modelID string) string {
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		return ""
	}
	switch strings.TrimSpace(runtimeProvider) {
	case "codex":
		if modelID == "codex-default" {
			return ""
		}
	case "claude", "claude_code":
		if modelID == "claude-default" {
			return ""
		}
	case "openclaw":
		if modelID == "openclaw-default" {
			return ""
		}
	case "cursor":
		if modelID == "cursor-auto" {
			return ""
		}
	}
	if modelID == "runtime-default" {
		return ""
	}
	return modelID
}

func eventRequestFromAgentEvent(assignment assignmentcontract.Assignment, ev agentbridge.Event) (assignmentcontract.AgentEventRequest, bool) {
	req := assignmentcontract.AgentEventRequest{
		AssignmentID: assignment.ID,
		TaskID:       assignment.TaskID,
	}
	switch ev.Kind {
	case agentbridge.EventProgress:
		req.EventType = assignmentcontract.EventRiidoLog
		req.Message = ev.Text
		req.Metadata = agentbridge.ProgressEventMetadata(ev)
	// NOTE: EventTextDelta is intentionally NOT forwarded. Providers (esp. codex)
	// emit deltas as tiny token/JSON fragments; surfacing each as its own progress
	// line produced incoherent, fragmented output ("code", "\":", "110", ...).
	// The control plane shows structured progress + the final result instead.
	// Coherent live streaming requires accumulating deltas into one evolving
	// message (a separate feature), not one progress line per delta.
	case agentbridge.EventLifecycle:
		if ev.Phase == agentbridge.StateRunning {
			req.EventType = assignmentcontract.EventAssignmentRunning
			req.State = assignmentcontract.AssignmentRunning
			req.Message = "provider running"
		} else {
			return req, false
		}
	case agentbridge.EventLog:
		req.EventType = assignmentcontract.EventProviderLog
		req.Message = ev.Text
	case agentbridge.EventWarning:
		req.EventType = assignmentcontract.EventProviderWarning
		req.Message = ev.Text
	case agentbridge.EventError:
		req.EventType = assignmentcontract.EventProviderError
		req.Message = textutil.FirstNonEmptyTrimmed(ev.Err, ev.Text)
	default:
		return req, false
	}
	return req, true
}

func terminalStateAndEvent(status agentbridge.ResultStatus) (assignmentcontract.AssignmentState, string) {
	switch status {
	case agentbridge.ResultCompleted:
		return assignmentcontract.AssignmentCompleted, assignmentcontract.EventAssignmentCompleted
	case agentbridge.ResultCancelled:
		return assignmentcontract.AssignmentCancelled, assignmentcontract.EventAssignmentCancelled
	default:
		return assignmentcontract.AssignmentFailed, assignmentcontract.EventAssignmentFailed
	}
}

func providerFromRuntimeID(runtimeID string) string {
	parts := strings.Split(runtimeID, ":")
	return strings.TrimSpace(parts[len(parts)-1])
}

func RuntimeIDForAgent(daemonID string, agent AgentBinding) string {
	return strings.TrimSpace(daemonID) + ":agent:" + url.QueryEscape(strings.TrimSpace(agent.AgentID)) + ":" + strings.TrimSpace(agent.RuntimeProvider)
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
		if agent.AgentID == assignment.AgentID && agent.RuntimeProvider == assignment.RuntimeProvider {
			return RuntimeIDForAgent(p.cfg.DaemonID, agent), nil
		}
	}
	return RuntimeIDForAgent(p.cfg.DaemonID, AgentBinding{AgentID: assignment.AgentID, RuntimeProvider: assignment.RuntimeProvider}), nil
}

func agentFromRuntimeID(runtimeID string) (string, bool) {
	parts := strings.Split(runtimeID, ":")
	if len(parts) < 4 || parts[len(parts)-3] != "agent" {
		return "", false
	}
	agentID, err := url.QueryUnescape(strings.TrimSpace(parts[len(parts)-2]))
	if err != nil {
		return "", false
	}
	agentID = strings.TrimSpace(agentID)
	return agentID, agentID != ""
}

func normalizeAgents(in []AgentBinding) []AgentBinding {
	out := make([]AgentBinding, 0, len(in))
	for _, agent := range in {
		agent.AgentID = strings.TrimSpace(agent.AgentID)
		agent.RuntimeProvider = strings.TrimSpace(agent.RuntimeProvider)
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

func assignmentExecutionID(assignment assignmentcontract.Assignment) string {
	return textutil.FirstNonEmptyTrimmed(assignment.ID, assignment.TaskID)
}

func sendAndCloseCancelWatcher(s *planeState, executionID string, cause error) {
	ch := s.cancelWatchers[executionID]
	if ch == nil {
		return
	}
	if cause != nil {
		select {
		case ch <- cause:
		default:
		}
	}
	closeCancelWatcher(s, executionID)
}

func closeCancelWatcher(s *planeState, executionID string) {
	ch := s.cancelWatchers[executionID]
	if ch == nil {
		return
	}
	close(ch)
	delete(s.cancelWatchers, executionID)
}
