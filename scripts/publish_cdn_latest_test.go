package scripts_test

import (
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
		"RIIDO_CDN_PREFIX=releases/latest/ai-agent",
		"RIIDO_CDN_DRY_RUN=true",
		"RIIDO_CLOUDFRONT_DISTRIBUTION_ID=TESTDIST",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dry-run sync failed: %v\n%s", err, out)
	}
	log := string(readFile(t, awsLog))
	if strings.Count(log, "s3 cp ") != 3 || !strings.Contains(log, "--dryrun") {
		t.Fatalf("unexpected aws dry-run log: %s", log)
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
