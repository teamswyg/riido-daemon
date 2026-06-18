package saasplane

import (
	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func assignmentResumeSessionID(assignment assignmentcontract.Assignment) string {
	return assignmentcontract.ResumeSessionIDForAssignment(assignment)
}

func cloneAssignmentWorktree(worktree *assignmentcontract.AssignmentWorktree) *assignmentcontract.AssignmentWorktree {
	if worktree == nil {
		return nil
	}
	out := *worktree
	return &out
}

func assignmentWorkspaceID(assignment assignmentcontract.Assignment) string {
	return textutil.FirstNonEmptyTrimmed(assignment.ComponentID, assignment.TaskID)
}
