package scripts_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestPublishCDNLatestRejectsReleaseTagVersionMismatch(t *testing.T) {
	dist := newCDNDistFixture(t, "v1.2.3")
	cmd := exec.Command("bash", "publish-cdn-latest.sh", "verify")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(),
		"RIIDO_CDN_DIST_DIR="+dist,
		"RIIDO_RELEASE_TAG=v9.9.9",
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected version mismatch failure, output=%s", out)
	}
	if !strings.Contains(string(out), "VERSION=v1.2.3 does not match RIIDO_RELEASE_TAG=v9.9.9") {
		t.Fatalf("unexpected mismatch output: %s", out)
	}
}

func TestPublishCDNLatestDryRunSyncWritesEvidenceWithoutInvalidation(t *testing.T) {
	dist := newCDNDistFixture(t, "v1.2.3")
	binDir := t.TempDir()
	awsLog := filepath.Join(t.TempDir(), "aws.log")
	evidence := filepath.Join(t.TempDir(), "cdn-evidence.json")
	writeExecutable(t, filepath.Join(binDir, "aws"), fakeAWSScript())
	cmd := exec.Command("bash", "publish-cdn-latest.sh", "sync", "-evidence-out", evidence)
	cmd.Dir = "."
	cmd.Env = append(os.Environ(),
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"AWS_LOG="+awsLog,
		"RIIDO_CDN_DIST_DIR="+dist,
		"RIIDO_RELEASE_TAG=v1.2.3",
		"RIIDO_CDN_BUCKET=riido-cdn-test",
		"RIIDO_CDN_PREFIX=",
		"RIIDO_CDN_DRY_RUN=true",
		"RIIDO_CLOUDFRONT_DISTRIBUTION_ID=TESTDIST",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dry-run sync failed: %v\n%s", err, out)
	}
	log := string(readFile(t, awsLog))
	if strings.Count(log, "s3 cp ") != 5 || !strings.Contains(log, "--dryrun") {
		t.Fatalf("unexpected aws dry-run log: %s", log)
	}
	for _, want := range []string{
		"riido-daemon_windows_amd64.zip",
		"riido-daemon_windows_arm64.zip",
	} {
		if !strings.Contains(log, want) {
			t.Fatalf("aws dry-run log missing %s: %s", want, log)
		}
	}
	if !strings.Contains(log, "s3://riido-cdn-test/releases/safe/ai-agent/") || strings.Contains(log, "releases/latest/") {
		t.Fatalf("unsafe CDN destination: %s", log)
	}
	if strings.Contains(log, "cloudfront create-invalidation") {
		t.Fatalf("dry-run must not invalidate CloudFront: %s", log)
	}
	got := string(readFile(t, evidence))
	for _, want := range []string{
		`"schema_version":"riido-cdn-latest-ai-agent.v1"`,
		`"release_tag":"v1.2.3"`,
		`"dry_run":"true"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("evidence missing %s: %s", want, got)
		}
	}
}

func TestPublishCDNRejectsBareMacPackageBeforeAnyUpload(t *testing.T) {
	dist := newCDNDistFixture(t, "v1.2.3")
	asset := "riido-daemon_darwin_arm64.tar.gz"
	source := t.TempDir()
	for _, name := range []string{"VERSION", "riido"} {
		content := "daemon"
		if name == "VERSION" {
			content = "v1.2.3\n"
		}
		if err := os.WriteFile(filepath.Join(source, name), []byte(content), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	archive := filepath.Join(dist, asset)
	tarCommand := exec.Command("tar", "-czf", archive, "-C", source, "VERSION", "riido")
	tarCommand.Env = append(os.Environ(), "COPYFILE_DISABLE=1")
	if out, err := tarCommand.CombinedOutput(); err != nil {
		t.Fatalf("%v %s", err, out)
	}
	sums := string(readFile(t, filepath.Join(dist, "SHA256SUMS")))
	lines := strings.Split(sums, "\n")
	for i, line := range lines {
		if strings.HasSuffix(line, "  "+asset) {
			lines[i] = fmt.Sprintf("%x  %s", sha256.Sum256(readFile(t, archive)), asset)
		}
	}
	if err := os.WriteFile(filepath.Join(dist, "SHA256SUMS"), []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	log := filepath.Join(t.TempDir(), "aws.log")
	writeExecutable(t, filepath.Join(bin, "aws"), fakeAWSScript())
	cmd := exec.Command("bash", "publish-cdn-latest.sh", "sync")
	cmd.Env = append(os.Environ(), "RIIDO_CDN_DIST_DIR="+dist, "RIIDO_RELEASE_TAG=v1.2.3", "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "AWS_LOG="+log)
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "release archive requires CLI auth helper") {
		t.Fatalf("%v %s", err, out)
	}
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatal("invalid release touched CDN")
	}
}
