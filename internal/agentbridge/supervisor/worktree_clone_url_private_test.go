package supervisor

import (
	"errors"
	"strings"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestAssignmentCloneURLRejectsPrivateWorktree(t *testing.T) {
	_, err := assignmentCloneURL(&assignmentcontract.AssignmentWorktree{
		RepositoryFullName: "teamswyg/private-repo",
		RepositoryURL:      "https://github.com/teamswyg/private-repo",
		IsPrivate:          true,
	})
	if err == nil || !strings.Contains(err.Error(), "private assignment worktree requires git credentials") {
		t.Fatalf("expected private repository fail-closed error, got %v", err)
	}
	if !errors.Is(err, errAssignmentWorktreeBlocked) {
		t.Fatalf("private repository error is not classified as blocked: %v", err)
	}
}
