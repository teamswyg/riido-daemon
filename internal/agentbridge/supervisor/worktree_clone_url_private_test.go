package supervisor

import (
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestAssignmentCloneURLAcceptsPrivateWorktree(t *testing.T) {
	got, err := assignmentCloneURL(&assignmentcontract.AssignmentWorktree{
		RepositoryFullName: "teamswyg/private-repo",
		RepositoryURL:      "https://github.com/teamswyg/private-repo",
		IsPrivate:          true,
	})
	if err != nil || got != "https://github.com/teamswyg/private-repo" {
		t.Fatalf("private repository URL = %q, error = %v", got, err)
	}
}

func TestAssignmentCloneURLRejectsSensitivePrivateURL(t *testing.T) {
	for _, tt := range unsupportedCloneURLCases() {
		t.Run(tt.name, func(t *testing.T) {
			_, err := assignmentCloneURL(&assignmentcontract.AssignmentWorktree{
				RepositoryFullName: tt.fullName, RepositoryURL: tt.repoURL, IsPrivate: true,
			})
			assertUnsupportedCloneURLError(t, err, tt.notContain)
		})
	}
}
