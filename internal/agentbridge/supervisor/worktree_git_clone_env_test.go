package supervisor

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestDefaultRunAssignmentGitClonePassesNonInteractiveEnv(t *testing.T) {
	if os.Getenv("RIIDO_TEST_GIT_CLONE_HELPER") == "1" {
		writeGitCloneHelperOutput(t)
		os.Exit(0)
	}

	outPath := t.TempDir() + "/git-env.txt"
	t.Setenv("RIIDO_TEST_GIT_CLONE_HELPER", "1")
	t.Setenv("RIIDO_TEST_GIT_CLONE_HELPER_OUT", outPath)
	t.Setenv("GIT_TERMINAL_PROMPT", "1")
	t.Setenv("GIT_ASKPASS", "must-not-launch")
	t.Setenv("GCM_INTERACTIVE", "Always")
	t.Setenv("RIIDO_DEVICE_SECRET", "fixture-daemon-secret")
	t.Setenv("LC_ALL", "ko_KR.UTF-8")

	err := defaultRunAssignmentGitClone(
		context.Background(),
		os.Args[0],
		gitCloneHelperArgs(),
	)
	if err != nil {
		t.Fatalf("defaultRunAssignmentGitClone: %v", err)
	}
	assertGitCloneHelperOutput(t, outPath)
}

func writeGitCloneHelperOutput(t *testing.T) {
	t.Helper()
	payload := strings.Join([]string{
		"prompt=" + os.Getenv("GIT_TERMINAL_PROMPT"),
		"gcm=" + os.Getenv("GCM_INTERACTIVE"),
		"askpass=" + os.Getenv("GIT_ASKPASS"),
		"daemon_secret=" + os.Getenv("RIIDO_DEVICE_SECRET"),
		"locale=" + os.Getenv("LC_ALL"),
		"path=" + os.Getenv("PATH"),
		"args=" + strings.Join(os.Args, "\n"),
	}, "\n")
	if err := os.WriteFile(os.Getenv("RIIDO_TEST_GIT_CLONE_HELPER_OUT"), []byte(payload), 0o600); err != nil {
		t.Fatalf("write helper output: %v", err)
	}
}

func gitCloneHelperArgs() []string {
	return []string{
		"-test.run=TestDefaultRunAssignmentGitClonePassesNonInteractiveEnv",
		"--",
		"clone",
		"--depth=1",
		"https://github.com/teamswyg/riido-daemon",
		"/tmp/riido-workdir",
	}
}

func assertGitCloneHelperOutput(t *testing.T, outPath string) {
	t.Helper()
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read helper output: %v", err)
	}
	got := string(raw)
	if !strings.Contains(got, "prompt=0") {
		t.Fatalf("GIT_TERMINAL_PROMPT was not disabled: %s", got)
	}
	for _, expected := range []string{"gcm=Never\n", "askpass=\n", "daemon_secret=\n", "locale=C\n"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("unsafe Git child environment: missing %q", expected)
		}
	}
	if !strings.Contains(got, "path=") || strings.Contains(got, "path=\n") {
		t.Fatalf("PATH was not inherited for git child process: %s", got)
	}
	for _, want := range []string{"clone", "--depth=1", "https://github.com/teamswyg/riido-daemon", "/tmp/riido-workdir"} {
		if !strings.Contains(got, want) {
			t.Fatalf("helper args missing %q: %s", want, got)
		}
	}
}
