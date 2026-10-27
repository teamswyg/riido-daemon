package saasplane

import (
	"context"
	"errors"
	"sync"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"
)

// executionClaimBinding is source-owned and process-local. This additional
// binding/lease copy is not exported through JSON, formatting, event metadata,
// or error messages; existing assignment/Ready contracts remain unchanged.
type executionClaimBinding struct {
	owner                         *Plane
	assignment                    assignmentcontract.Assignment
	runtimeID, deviceID, daemonID string
	mu                            sync.Mutex
	terminal                      *assignmentcontract.AgentEventRequest
	terminalACK                   bool
	proof                         executionproof.Proof
	known                         bool
}

func (*executionClaimBinding) String() string   { return "[redacted execution claim]" }
func (*executionClaimBinding) GoString() string { return "[redacted execution claim]" }

func (p *Plane) claimBinding(ctx context.Context, executionID string) (*executionClaimBinding, error) {
	scope := executionproof.ScopeFromContext(ctx)
	if scope == nil || scope.Source() == nil {
		return nil, nil
	}
	binding, ok := scope.Source().(*executionClaimBinding)
	if !ok || binding.owner != p || assignmentExecutionID(binding.assignment) != executionID {
		return nil, errors.New("execution claim identity mismatch")
	}
	return binding, nil
}

func (p *Plane) saveClaimRuntime(ctx context.Context, a assignmentcontract.Assignment, runtimeID string, fresh bool) (*executionproof.Scope, error) {
	var scope *executionproof.Scope
	err := p.withState(ctx, func(s *planeState) {
		executionID := assignmentExecutionID(a)
		scope = s.executionScopes[executionID]
		previous, _ := scope.Source().(*executionClaimBinding)
		if previous == nil || previous.deviceID != p.cfg.DeviceID || previous.assignment.ID != a.ID || previous.assignment.LeaseToken != a.LeaseToken {
			// Capture only immutable identity fields; mutable provider metadata and
			// prompt/worktree pointers must never become evidence authority.
			identity := assignmentcontract.Assignment{ID: a.ID, TaskID: a.TaskID, ComponentID: a.ComponentID, AgentID: a.AgentID, RuntimeProvider: a.RuntimeProvider, LeaseToken: a.LeaseToken}
			binding := &executionClaimBinding{owner: p, assignment: identity, runtimeID: runtimeID, deviceID: p.cfg.DeviceID, daemonID: p.cfg.DaemonID}
			covered := fresh && validTerminationHeader(p.cfg.DeviceID, 512) && validTerminationHeader(p.cfg.DeviceSecret, 8192) && validTerminationHeader(a.LeaseToken, 8192) && a.ID != ""
			scope = executionproof.NewScope(covered, binding)
			s.executionScopes[executionID] = scope
			// The text buffer belongs to this claim, never to a later lease epoch.
			s.partialBodies[executionID] = &partialBodyState{claimedAt: time.Now()}
		}
		s.assignmentsByExecution[executionID] = a
		if runtimeID != "" {
			s.runtimeIDsByExecution[executionID] = runtimeID
		}
	})
	return scope, err
}

func currentClaimScope(ctx context.Context, s *planeState, executionID string) bool {
	scope := executionproof.ScopeFromContext(ctx)
	if scope == nil || scope.Source() == nil {
		return true
	}
	return s.executionScopes[executionID] == scope
}
