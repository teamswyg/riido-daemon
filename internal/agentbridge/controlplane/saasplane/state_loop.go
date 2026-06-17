package saasplane

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func (p *Plane) activeAssignmentsByAgentForHeartbeat(ctx context.Context, runningTaskIDs []string) (map[string][]string, error) {
	executions := normalizedExecutionIDs(runningTaskIDs)
	if len(executions) == 0 {
		return nil, nil
	}
	byAgent := map[string][]string{}
	err := p.withState(ctx, func(s *planeState) {
		for _, executionID := range executions {
			assignment := s.assignmentsByExecution[executionID]
			if assignment.AgentID == "" || assignment.ID == "" {
				continue
			}
			byAgent[assignment.AgentID] = append(byAgent[assignment.AgentID], assignment.ID)
		}
	})
	return byAgent, err
}

func (p *Plane) deliverCancel(ctx context.Context, assignment assignmentcontract.Assignment) error {
	return p.withState(ctx, func(s *planeState) {
		sendAndCloseCancelWatcher(s, assignmentExecutionID(assignment), fmt.Errorf("saas assignment %s cancelled", assignment.ID))
	})
}

func (p *Plane) deliverUnrefreshedHeartbeatCancels(ctx context.Context, requestedAssignmentIDs []string, response assignmentcontract.AgentHeartbeatResponse) error {
	if len(requestedAssignmentIDs) == 0 {
		return nil
	}
	refreshed := map[string]bool{}
	for _, assignment := range response.RefreshedAssignments {
		if strings.TrimSpace(assignment.ID) != "" {
			refreshed[assignment.ID] = true
		}
	}
	return p.withState(ctx, func(s *planeState) {
		for _, assignmentID := range requestedAssignmentIDs {
			assignmentID = strings.TrimSpace(assignmentID)
			if assignmentID == "" || refreshed[assignmentID] {
				continue
			}
			if assignment, ok := s.assignmentsByExecution[assignmentID]; ok {
				sendAndCloseCancelWatcher(s, assignmentID, fmt.Errorf("saas assignment %s heartbeat lease stale", assignment.ID))
				delete(s.assignmentsByExecution, assignmentID)
				delete(s.runtimeIDsByExecution, assignmentID)
				delete(s.partialBodies, assignmentID)
			}
		}
	})
}

type planeState struct {
	assignmentsByExecution  map[string]assignmentcontract.Assignment
	runtimeIDsByExecution   map[string]string
	cancelWatchers          map[string]chan error
	registeredRuntimes      map[string]RuntimeSnapshotRecord
	registeredDeviceName    string
	lastRuntimeSnapshotSync time.Time
	agentBindingsCache      []assignmentcontract.AgentRuntimeBinding
	agentBindingsCachedAt   time.Time
	nextAssignmentEventSeq  uint64
	// partialBodies accumulates each execution's assistant text deltas between
	// flushes so the daemon can forward a coherent evolving body instead of
	// per-token fragments. Keyed by execution ID.
	partialBodies map[string]*partialBodyState
}

// partialBodyState holds the running assistant text for one task and the
// debounce bookkeeping for forwarding it as an evolving progress line.
type partialBodyState struct {
	text           string
	lastFlushAt    time.Time
	lastFlushedLen int
}

type stateOp struct {
	fn    func(*planeState)
	close bool
	ack   chan struct{}
}

func (p *Plane) loop() {
	state := planeState{
		assignmentsByExecution: map[string]assignmentcontract.Assignment{},
		runtimeIDsByExecution:  map[string]string{},
		cancelWatchers:         map[string]chan error{},
		registeredRuntimes:     map[string]RuntimeSnapshotRecord{},
		partialBodies:          map[string]*partialBodyState{},
	}
	defer close(p.done)
	for op := range p.ops {
		if op.close {
			for _, ch := range state.cancelWatchers {
				close(ch)
			}
			close(op.ack)
			return
		}
		op.fn(&state)
		close(op.ack)
	}
}

func (p *Plane) withState(ctx context.Context, fn func(*planeState)) error {
	ack := make(chan struct{})
	select {
	case p.ops <- stateOp{fn: fn, ack: ack}:
	case <-p.done:
		return errors.New("saasplane: closed")
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case <-ack:
		return nil
	case <-p.done:
		return errors.New("saasplane: closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func taskRequestFromAssignment(assignment assignmentcontract.Assignment) *bridge.TaskRequest {
	executionID := assignmentExecutionID(assignment)
	metadata := map[string]string{
		MetadataAssignmentID:        assignment.ID,
		MetadataAgentID:             assignment.AgentID,
		MetadataComponentID:         assignment.ComponentID,
		MetadataLeaseToken:          assignment.LeaseToken,
		MetadataModelID:             assignment.ModelID,
		MetadataRuntimeProvider:     assignment.RuntimeProvider,
		controlplane.MetadataTaskID: assignment.TaskID,
		"workspace_id":              textutil.FirstNonEmptyTrimmed(assignment.ComponentID, assignment.TaskID),
		"run_id":                    executionID,
	}
	prompt, systemPrompt, telemetryPlacement, instructionPlacement := agentbridge.ApplyRuntimeInstructionContract(assignment.RuntimeProvider, assignment.Prompt, "", assignment.AgentInstruction)
	metadata[agentbridge.MetadataTelemetryContract] = telemetryPlacement
	if instructionPlacement != "" {
		metadata[agentbridge.MetadataAgentInstruction] = instructionPlacement
	}
	return &bridge.TaskRequest{
		ID:                       executionID,
		Provider:                 bridge.Provider(assignment.RuntimeProvider),
		Model:                    providercatalog.ModelOverride(assignment.RuntimeProvider, assignment.ModelID),
		Prompt:                   prompt,
		SystemPrompt:             systemPrompt,
		AllowExperimentalRuntime: assignment.AllowExperimentalRuntime,
		ResumeSessionID:          assignmentResumeSessionID(assignment),
		Worktree:                 cloneAssignmentWorktree(assignment.Worktree),
		Metadata:                 metadata,
	}
}

func assignmentResumeSessionID(assignment assignmentcontract.Assignment) string {
	return textutil.FirstNonEmptyTrimmed(assignment.ProviderSessionID, assignment.ResumeSessionID)
}

func cloneAssignmentWorktree(worktree *assignmentcontract.AssignmentWorktree) *assignmentcontract.AssignmentWorktree {
	if worktree == nil {
		return nil
	}
	out := *worktree
	return &out
}
