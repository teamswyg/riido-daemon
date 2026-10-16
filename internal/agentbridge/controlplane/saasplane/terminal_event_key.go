package saasplane

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

// Stable across both HTTP retries and supervisor report retries. Scope by lease
// so a replay from another ownership period cannot acknowledge a new run.
func terminalEventKey(assignment assignmentcontract.Assignment, executionID string, request assignmentcontract.AgentEventRequest) string {
	// AgentEventRequest consists solely of JSON-safe scalar fields and a string map.
	body, _ := json.Marshal(request) //nolint:errchkjson // All request fields are JSON-safe scalars or map[string]string.
	sum := sha256.Sum256(append([]byte(executionID+"\x00"+assignment.LeaseToken+"\x00"), body...))
	return fmt.Sprintf("terminal:%x", sum[:])
}
