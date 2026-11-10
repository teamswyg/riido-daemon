package supervisor

import (
	"context"
	"errors"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func TestPrivateWorktreeUsesExistingGitCredentialHelper(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is required for the authenticated clone fixture")
	}
	root := t.TempDir()
	config := filepath.Join(root, "gitconfig")
	if err := os.WriteFile(config, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", config)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_COUNT", "0")
	t.Setenv("GIT_ASKPASS", "must-not-run-an-interactive-prompt")
	seed := filepath.Join(root, "seed")
	privateFixtureGit(t, git, "init", "--initial-branch=main", seed)
	if err := os.WriteFile(filepath.Join(seed, "README.md"), []byte("private fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	privateFixtureGit(t, git, "-C", seed, "add", "README.md")
	privateFixtureGit(t, git, "-C", seed, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "commit", "-m", "fixture")
	privateFixtureGit(t, git, "clone", "--bare", seed, filepath.Join(root, "fixture.git"))
	backend := &cgi.Handler{Path: git, Args: []string{"http-backend"}, Env: []string{"GIT_PROJECT_ROOT=" + root, "GIT_HTTP_EXPORT_ALL=1"}}
	var authenticated atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || user != "fixture" || password != "fixture-password" {
			w.Header().Set("WWW-Authenticate", `Basic realm="fixture"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		authenticated.Add(1)
		backend.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)
	privateFixtureGit(t, git, "config", "--file", config, "url."+server.URL+"/fixture.git.insteadOf", "https://github.com/fixture/private-repo")
	worktree := &assignmentcontract.AssignmentWorktree{RepositoryURL: "https://github.com/fixture/private-repo", BranchName: "main", IsPrivate: true}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	err = materializeAssignmentWorktree(ctx, filepath.Join(root, "denied"), worktree)
	if !errors.Is(err, errAssignmentWorktreeBlocked) {
		t.Fatalf("missing credentials must block the private clone, got %v", err)
	}
	if !strings.Contains(err.Error(), "Git authentication is required") || !strings.Contains(err.Error(), "git ls-remote") {
		t.Fatalf("missing credentials must explain the cause and recovery: %v", err)
	}
	privateFixtureGit(t, git, "config", "--file", config, "credential.helper", `!f() { printf 'username=fixture\npassword=fixture-wrong-password\n'; }; f`)
	err = materializeAssignmentWorktree(ctx, filepath.Join(root, "rejected"), worktree)
	if !errors.Is(err, errAssignmentWorktreeBlocked) || !strings.Contains(err.Error(), "Git authentication was rejected") || strings.Contains(err.Error(), "fixture-wrong-password") {
		t.Fatalf("rejected credentials must explain authentication recovery without exposing credentials: %v", err)
	}
	// The helper has synthetic credentials, is configured by the host, and is
	// never read by the daemon. Git invokes it through its normal credential API.
	privateFixtureGit(t, git, "config", "--file", config, "credential.helper", `!f() { printf 'username=fixture\npassword=fixture-password\n'; }; f`)
	target := filepath.Join(root, "allowed")
	if err := materializeAssignmentWorktree(ctx, target, worktree); err != nil {
		t.Fatalf("authenticated private clone: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(target, "README.md"))
	if err != nil || string(content) != "private fixture\n" || authenticated.Load() == 0 {
		t.Fatalf("authenticated clone did not materialize the repository: %v", err)
	}
	origin := privateFixtureGit(t, git, "-C", target, "remote", "get-url", "origin")
	if strings.Contains(origin, "fixture-password") {
		t.Fatal("credential persisted in clone origin")
	}
	worktree.BranchName = "missing-branch"
	err = materializeAssignmentWorktree(ctx, filepath.Join(root, "missing-branch"), worktree)
	if !errors.Is(err, errAssignmentWorktreeBlocked) || !strings.Contains(err.Error(), "selected branch was not found") {
		t.Fatalf("missing branch must not be reported as an authentication failure: %v", err)
	}
}

func privateFixtureGit(t *testing.T, git string, args ...string) string {
	t.Helper()
	out, err := exec.CommandContext(t.Context(), git, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("fixture Git command failed: %v: %s", err, out)
	}
	return string(out)
}
