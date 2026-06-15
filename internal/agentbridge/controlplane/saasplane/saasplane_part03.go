package saasplane

import (
	"context"
	"errors"
	"strings"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func runtimeSnapshotFromHeartbeat(hb controlplane.RuntimeHeartbeat) (RuntimeSnapshotRecord, bool) {
	runtimeID := strings.TrimSpace(hb.RuntimeID)
	provider := providerFromRuntimeID(runtimeID)
	if runtimeID == "" || provider == "" {
		return RuntimeSnapshotRecord{}, false
	}
	return RuntimeSnapshotRecord{
		RuntimeID:      runtimeID,
		Kind:           runtimeKindForProvider(provider),
		Availability:   "online",
		DetectionState: "detected",
	}, true
}

func (p *Plane) postRuntimeSnapshot(ctx context.Context, runtimes []RuntimeSnapshotRecord, deviceName string) error {
	var out struct {
		SchemaVersion string `json:"schema_version"`
	}
	return p.postJSON(ctx, "/v1/daemon/runtime-snapshot", DeviceRuntimeSnapshotSyncRequest{
		DaemonID:          p.cfg.DaemonID,
		DeviceID:          p.cfg.DeviceID,
		DeviceDisplayName: textutil.FirstNonEmptyTrimmed(deviceName, p.cfg.DeviceID),
		Profile:           p.cfg.Profile,
		AppVersion:        p.cfg.AppVersion,
		PID:               p.cfg.PID,
		UptimeSeconds:     p.daemonUptimeSeconds(),
		StartedAt:         p.cfg.StartedAt,
		Runtimes:          runtimes,
	}, &out)
}

func (p *Plane) daemonUptimeSeconds() int64 {
	if p.cfg.StartedAt.IsZero() {
		return 0
	}
	seconds := int64(time.Since(p.cfg.StartedAt).Seconds())
	if seconds < 0 {
		return 0
	}
	return seconds
}

func (p *Plane) ClaimTask(ctx context.Context, runtimeID string) (*bridge.TaskRequest, error) {
	provider := providerFromRuntimeID(runtimeID)
	if p.dynamicBindingsEnabled() {
		bindings, err := p.agentBindings(ctx)
		if err != nil {
			return nil, err
		}
		for _, binding := range bindings {
			if binding.RuntimeProvider != provider || strings.TrimSpace(binding.RuntimeID) != strings.TrimSpace(runtimeID) {
				continue
			}
			poll, err := p.pollAgent(ctx, binding.AgentID, runtimeID)
			if err != nil {
				return nil, err
			}
			if poll.Assignment == nil {
				continue
			}
			switch poll.Action {
			case assignmentcontract.PollStart, assignmentcontract.PollActive:
				assignment := *poll.Assignment
				if assignment.RuntimeProvider != "" && assignment.RuntimeProvider != provider {
					continue
				}
				if err := p.saveAssignmentRuntime(ctx, assignment, runtimeID); err != nil {
					return nil, err
				}
				return taskRequestFromAssignment(assignment), nil
			case assignmentcontract.PollCancel:
				_ = p.deliverCancel(ctx, *poll.Assignment)
				return nil, nil
			case assignmentcontract.PollNone:
				continue
			default:
				continue
			}
		}
		return nil, nil
	}
	runtimeAgent, hasRuntimeAgent := agentFromRuntimeID(runtimeID)
	for _, agent := range p.cfg.Agents {
		if agent.RuntimeProvider != provider {
			continue
		}
		if hasRuntimeAgent && agent.AgentID != runtimeAgent {
			continue
		}
		poll, err := p.pollAgent(ctx, agent.AgentID, runtimeID)
		if err != nil {
			return nil, err
		}
		if poll.Assignment == nil {
			continue
		}
		switch poll.Action {
		case assignmentcontract.PollStart, assignmentcontract.PollActive:
			assignment := *poll.Assignment
			if assignment.RuntimeProvider != "" && assignment.RuntimeProvider != provider {
				continue
			}
			if err := p.saveAssignmentRuntime(ctx, assignment, runtimeID); err != nil {
				return nil, err
			}
			return taskRequestFromAssignment(assignment), nil
		case assignmentcontract.PollCancel:
			_ = p.deliverCancel(ctx, *poll.Assignment)
			return nil, nil
		case assignmentcontract.PollNone:
			continue
		default:
			continue
		}
	}
	return nil, nil
}

func (p *Plane) WatchCancellation(ctx context.Context, executionID string) (<-chan error, error) {
	executionID = strings.TrimSpace(executionID)
	if executionID == "" {
		return nil, errors.New("saasplane: empty executionID")
	}
	ch := make(chan error, 1)
	err := p.withState(ctx, func(s *planeState) {
		closeCancelWatcher(s, executionID)
		s.cancelWatchers[executionID] = ch
	})
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (p *Plane) StartTask(ctx context.Context, executionID string) error {
	assignment, ok, err := p.assignmentForExecution(ctx, executionID)
	if err != nil || !ok {
		return err
	}
	_, err = p.postAgentEvent(ctx, assignment, assignmentcontract.AgentEventRequest{
		AssignmentID: assignment.ID,
		TaskID:       assignment.TaskID,
		State:        assignmentcontract.AssignmentReady,
		EventType:    assignmentcontract.EventAssignmentReady,
		Message:      "daemon ready",
	})
	return err
}

func (p *Plane) ReportEvent(ctx context.Context, executionID string, ev agentbridge.Event) error {
	assignment, ok, err := p.assignmentForExecution(ctx, executionID)
	if err != nil || !ok {
		return err
	}
	if ev.Kind == agentbridge.EventTextDelta {
		return p.accumulatePartialBody(ctx, assignment, executionID, ev.Text)
	}
	req, ok := eventRequestFromAgentEvent(assignment, ev)
	if !ok {
		return nil
	}
	_, err = p.postAgentEvent(ctx, assignment, req)
	return err
}
