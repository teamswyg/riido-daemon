package supervisor

import (
	"context"
	"errors"
	"reflect"
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
}

func TestAssignmentCloneURLRejectsUnsupportedRepositoryURL(t *testing.T) {
	tests := []struct {
		name       string
		repoURL    string
		fullName   string
		notContain []string
	}{
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
		{
			name:    "empty force query",
			repoURL: "https://github.com/teamswyg/riido-daemon?",
		},
		{
			name:       "fragment token",
			repoURL:    "https://github.com/teamswyg/riido-daemon#secret-token",
			notContain: []string{"secret-token"},
		},
		{
			name:    "missing repo path",
			repoURL: "https://github.com",
		},
		{
			name:     "full name query",
			fullName: "teamswyg/riido-daemon?token=secret",
			notContain: []string{
				"secret",
				"token=",
			},
		},
		{
			name:     "full name encoded query",
			fullName: "teamswyg/riido-daemon%3Ftoken=secret",
			notContain: []string{
				"secret",
				"token=",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := assignmentCloneURL(&assignmentcontract.AssignmentWorktree{
				RepositoryFullName: tt.fullName,
				RepositoryURL:      tt.repoURL,
			})
			if err == nil {
				t.Fatal("expected unsupported URL error")
			}
			for _, forbidden := range tt.notContain {
				if strings.Contains(err.Error(), forbidden) {
					t.Fatalf("error leaked sensitive URL component %q: %v", forbidden, err)
				}
			}
		})
	}
}

func TestMaterializeAssignmentWorktreeRunsShallowBranchClone(t *testing.T) {
	var gotGit string
	var gotArgs []string
	original := runAssignmentGitClone
	runAssignmentGitClone = func(_ context.Context, git string, args []string) error {
		gotGit = git
		gotArgs = append([]string(nil), args...)
		return nil
	}
	t.Cleanup(func() { runAssignmentGitClone = original })

	err := materializeAssignmentWorktree(context.Background(), "/tmp/riido-workdir", &assignmentcontract.AssignmentWorktree{
		RepositoryFullName: "teamswyg/riido-daemon",
		RepositoryURL:      "https://github.com/teamswyg/riido-daemon",
		BranchName:         "RIID-4964-agent-profile-upload",
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "git executable not found") {
			t.Skipf("git not available in test environment: %v", err)
		}
		t.Fatalf("materialize worktree: %v", err)
	}
	if gotGit == "" {
		t.Fatal("git executable was not resolved")
	}
	wantArgs := []string{
		"clone",
		"--depth=1",
		"--branch",
		"RIID-4964-agent-profile-upload",
		"https://github.com/teamswyg/riido-daemon",
		"/tmp/riido-workdir",
	}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("clone args = %#v, want %#v", gotArgs, wantArgs)
	}
}
