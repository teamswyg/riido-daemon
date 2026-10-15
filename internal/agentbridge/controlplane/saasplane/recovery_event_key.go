package saasplane

import (
	"crypto/sha256"
	"fmt"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func recoveryEventKey(assignment assignmentcontract.Assignment, executionID string, c assignmentcontract.RecoveryContext) string {
	c = c.Sanitized()
	sum := sha256.Sum256(fmt.Appendf(nil, "%s\x00%s\x00%d\x00%s\x00%s\x00%d\x00%s\x00%s\x00%s", assignment.ID, executionID, c.Attempt, c.State, c.Provider, c.SwitchCount, c.Reason, c.Circuit, c.OutcomeCertainty))
	return fmt.Sprintf("recovery:%x", sum[:])
}
