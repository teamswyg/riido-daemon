package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A reused package directory must never leak optional files into a legacy build.
func TestReleasePackageIgnoresStaleHelper(t *testing.T) {
	root := newReleaseBuildFixture(t)
	stage := filepath.Join(root, "temp", "package-linux-amd64")
	if err := os.MkdirAll(stage, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stage, "CLI_AUTH_HELPER.json"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := releaseBuildCommand(t, root, "linux/amd64/tar")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, out)
	}
	list := exec.Command("tar", "-tzf", filepath.Join(root, "dist/linux_amd64/riido-daemon_linux_amd64.tar.gz"))
	out, err := list.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "CLI_AUTH_HELPER") {
		t.Fatalf("stale helper leaked: %s", out)
	}
}

// Explicit helper opt-in must fail closed, preserving any previous archive.
func TestReleasePackageRejectsUnsupportedHelperWithoutReplacingArchive(t *testing.T) {
	root := newReleaseBuildFixture(t)
	archive := filepath.Join(root, "dist/linux_amd64/riido-daemon_linux_amd64.tar.gz")
	if err := os.MkdirAll(filepath.Dir(archive), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(archive, []byte("previous-asset"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := releaseBuildCommand(t, root, "linux/amd64/tar")
	cmd.Env = append(cmd.Env, "RIIDO_CLI_AUTH_HELPER_BUNDLE="+filepath.Join(root, "missing.app"))
	if out, err := cmd.CombinedOutput(); err == nil {
		t.Fatalf("unsupported helper accepted: %s", out)
	}
	if got := string(readFile(t, archive)); got != "previous-asset" {
		t.Fatalf("prior asset replaced: %q", got)
	}
}

func newReleaseBuildFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, name := range []string{"temp", "bin"} {
		if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"LICENSE", "NOTICE.md"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeExecutable(t, filepath.Join(root, "bin/go"), `#!/bin/sh
while [ "$#" -gt 0 ]; do
 if [ "$1" = "-o" ]; then shift; printf 'fixture binary\n' > "$1"; chmod +x "$1"; exit 0; fi
 shift
done
exit 1
`)
	return root
}

func releaseBuildCommand(t *testing.T, root, target string) *exec.Cmd {
	t.Helper()
	script, err := filepath.Abs("build-release-package.sh")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", script, target)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "PATH="+filepath.Join(root, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"), "GITHUB_REF_NAME=v1.2.3", "RUNNER_TEMP="+filepath.Join(root, "temp"), "GITHUB_WORKSPACE="+root, "GITHUB_ENV="+filepath.Join(root, "github.env"))
	return cmd
}

func TestReleasePackageSupportsGitHubPullRequestRef(t *testing.T) {
	root := newReleaseBuildFixture(t)
	cmd := releaseBuildCommand(t, root, "linux/amd64/tar")
	cmd.Env = append(cmd.Env, "GITHUB_EVENT_NAME=pull_request", "GITHUB_REF_NAME=759/merge")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("PR package failed: %v %s", err, out)
	}
	readVersion := exec.Command("tar", "-xOzf", filepath.Join(root, "dist/linux_amd64/riido-daemon_linux_amd64.tar.gz"), "./VERSION")
	data, err := readVersion.CombinedOutput()
	if err != nil || string(data) != "pull-request\n" {
		t.Fatalf("PR VERSION=%q error=%v", data, err)
	}
}
