package supervisor

import (
	"errors"
	"strings"
	"testing"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

type unsupportedCloneURLCase struct {
	name       string
	repoURL    string
	fullName   string
	notContain []string
}

func TestAssignmentCloneURLRejectsUnsupportedRepositoryURL(t *testing.T) {
	for _, tt := range unsupportedCloneURLCases() {
		t.Run(tt.name, func(t *testing.T) {
			_, err := assignmentCloneURL(&assignmentcontract.AssignmentWorktree{
				RepositoryFullName: tt.fullName,
				RepositoryURL:      tt.repoURL,
			})
			assertUnsupportedCloneURLError(t, err, tt.notContain)
		})
	}
}

func unsupportedCloneURLCases() []unsupportedCloneURLCase {
	return []unsupportedCloneURLCase{
		{
			name:       "userinfo",
			repoURL:    "https://token:secret@example.com/teamswyg/riido-daemon",
			notContain: []string{"secret", "token"},
		},
		{
			name:       "query token",
			repoURL:    "https://github.com/teamswyg/riido-daemon?token=secret",
			notContain: []string{"secret", "token="},
		},
		{name: "empty force query", repoURL: "https://github.com/teamswyg/riido-daemon?"},
		{
			name:       "fragment token",
			repoURL:    "https://github.com/teamswyg/riido-daemon#secret-token",
			notContain: []string{"secret-token"},
		},
		{name: "missing repo path", repoURL: "https://github.com"},
		{
			name:       "full name query",
			fullName:   "teamswyg/riido-daemon?token=secret",
			notContain: []string{"secret", "token="},
		},
		{
			name:       "full name encoded query",
			fullName:   "teamswyg/riido-daemon%3Ftoken=secret",
			notContain: []string{"secret", "token="},
		},
	}
}

func assertUnsupportedCloneURLError(t *testing.T, err error, forbidden []string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected unsupported URL error")
	}
	if !errors.Is(err, errAssignmentWorktreeBlocked) {
		t.Fatalf("unsupported repository error is not classified as blocked: %v", err)
	}
	for _, token := range forbidden {
		if strings.Contains(err.Error(), token) {
			t.Fatalf("error leaked sensitive URL component %q: %v", token, err)
		}
	}
}
