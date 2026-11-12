package supervisor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestPrivateWorktreeReportsCauseAndRecovery(t *testing.T) {
	if os.Getenv("RIIDO_TEST_GIT_DIAGNOSTIC_HELPER") == "1" {
		fmt.Fprintln(os.Stdout, "fixture-secret-stdout")
		fmt.Fprintln(os.Stderr, os.Getenv("RIIDO_TEST_GIT_STDERR"))
		os.Exit(1)
	}
	t.Setenv("RIIDO_TEST_GIT_DIAGNOSTIC_HELPER", "1")
	originalClone, originalResolve := runAssignmentGitClone, resolveAssignmentGitExecutable
	t.Cleanup(func() {
		runAssignmentGitClone, resolveAssignmentGitExecutable = originalClone, originalResolve
	})
	resolveAssignmentGitExecutable = func(string, string) (string, bool) { return os.Args[0], true }
	runAssignmentGitClone = func(ctx context.Context, git string, _ []string) error {
		return defaultRunAssignmentGitClone(ctx, git, []string{"-test.run=^TestPrivateWorktreeReportsCauseAndRecovery$"})
	}
	for _, tt := range []struct{ name, stderr, reason, action string }{
		{"missing credentials", "fatal: could not read Username for 'https://fixture-secret@github.com': terminal prompts disabled", "Git authentication is required", "gh auth setup-git --hostname github.com"},
		{"noninteractive manager", "fatal: Cannot prompt because user interactivity has been disabled", "Git authentication is required", "git ls-remote"},
		{"rejected credentials", "fatal: Authentication failed for 'https://fixture-secret@github.com/repo'", "Git authentication was rejected", "Refresh the saved Git credentials"},
		{"invalid token", "remote: Invalid username or token", "Git authentication was rejected", "gh auth login"},
		{"unauthorized", "fatal: The requested URL returned error: 401", "Git authentication was rejected", "Refresh the saved Git credentials"},
		{"repository hidden", "remote: Repository not found.", "account may lack access", "Verify the repository URL"},
		{"forbidden", "fatal: The requested URL returned error: 403", "account may lack access", "authorize organization SSO"},
		{"missing branch", "fatal: Remote branch fixture-secret not found in upstream origin", "selected branch was not found", "push the intended branch"},
		{"missing ref", "fatal: couldn't find remote ref fixture-secret", "selected branch was not found", "Select an existing branch"},
		{"dns", "fatal: unable to access 'https://fixture-secret@github.com': Could not resolve host: github.com", "could not connect", "DNS, VPN, proxy"},
		{"timeout", "fatal: Failed to connect to github.com: Connection timed out", "could not connect", "firewall settings"},
		{"certificate", "fatal: SSL certificate problem: unable to get local issuer certificate", "could not verify the HTTPS certificate", "required CA certificate"},
		{"unclassified", "fatal: unexpected fixture-secret-helper-error", "exact cause could not be determined", "git ls-remote"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("RIIDO_TEST_GIT_STDERR", tt.stderr+"\nfixture-secret-token")
			err := materializeAssignmentWorktree(t.Context(), t.TempDir(), privateWorktreeRequest().Worktree, nil)
			if !errors.Is(err, errAssignmentWorktreeBlocked) {
				t.Fatalf("expected repository preparation failure, got %v", err)
			}
			for _, want := range []string{"Reason:", tt.reason, "Action:", tt.action} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("diagnostic must include %q: %v", want, err)
				}
			}
			if strings.Contains(err.Error(), "fixture-secret") {
				t.Fatal("raw Git/helper output leaked into the task diagnostic")
			}
		})
	}
}

func TestGitDiagnosticRetainsBoundedTail(t *testing.T) {
	var buffer assignmentGitDiagnosticBuffer
	for _, chunk := range []string{strings.Repeat("x", assignmentGitDiagnosticLimit*2), "fatal: Authentication failed"} {
		n, err := buffer.Write([]byte(chunk))
		if err != nil || n != len(chunk) || len(buffer.data) > assignmentGitDiagnosticLimit {
			t.Fatal("diagnostic buffer must consume all bytes and retain only the bounded tail")
		}
	}
	if err := classifyAssignmentGitCloneFailure(buffer.data); !strings.Contains(err.Error(), "Git authentication was rejected") {
		t.Fatalf("final Git error was lost after verbose output: %v", err)
	}
}
