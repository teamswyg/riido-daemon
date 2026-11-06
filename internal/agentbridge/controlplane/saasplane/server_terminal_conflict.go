package saasplane

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

var errServerTerminal = errors.New("assignment report superseded by terminal server state")

// RecordAgentEvent validates the addressed assignment and returns its stored
// state in the transition error. No second read (or speculative poll) is needed.
func confirmedTerminalConflict(err error, requested assignmentcontract.AssignmentState) (assignmentcontract.AssignmentState, bool) {
	var status httpStatusError
	if requested == "" || !errors.As(err, &status) || status.StatusCode != http.StatusBadRequest {
		return "", false
	}
	var response struct {
		SchemaVersion string `json:"schema_version"`
		Error         string `json:"error"`
	}
	if json.Unmarshal([]byte(status.Body), &response) != nil || response.SchemaVersion != assignmentcontract.SchemaVersion {
		return "", false
	}
	transition, ok := strings.CutPrefix(response.Error, "invalid assignment transition ")
	from, to, okPair := strings.Cut(transition, " -> ")
	state := assignmentcontract.AssignmentState(from)
	return state, ok && okPair && assignmentcontract.IsTerminal(state) && state != requested && to == string(requested)
}

func (p *Plane) retireObsoleteReport(ctx context.Context, assignment assignmentcontract.Assignment, binding *executionClaimBinding, serverState, requested assignmentcontract.AssignmentState) error {
	executionID := assignmentExecutionID(assignment)
	err := p.withState(ctx, func(s *planeState) {
		current, ok := s.assignmentsByExecution[executionID]
		if !ok || !currentClaimScope(ctx, s, executionID) || current.ID != assignment.ID || current.LeaseToken != assignment.LeaseToken {
			return
		}
		closeCancelWatcher(s, executionID)
		delete(s.assignmentsByExecution, executionID)
		delete(s.runtimeIDsByExecution, executionID)
		delete(s.partialBodies, executionID)
	})
	if err != nil {
		return err
	}
	if binding == nil || binding.serverTerminal.CompareAndSwap(false, true) {
		ref := sha256.Sum256([]byte(assignment.ID))
		log.Printf("riido_daemon event=assignment_report_superseded assignment_ref=%x server_state=%q requested_state=%q", ref[:8], serverState, requested)
	}
	return errServerTerminal
}
