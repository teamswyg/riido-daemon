package main

import (
	"strings"
	"testing"
)

func TestRenderPlistIncludesDailySchedule(t *testing.T) {
	cfg := testConfig()
	paths := schedulePaths{repo: "/tmp/repo", stdout: "/tmp/out", stderr: "/tmp/err"}
	got := renderPlist(cfg, paths)
	for _, want := range []string{
		"<key>Hour</key><integer>9</integer>",
		"<key>Minute</key><integer>5</integer>",
		"go run ./tools/localqarunner",
		"s3://bucket/daily",
		"/tmp/product.json",
		"-coverage-evidence &#39;/tmp/coverage.json&#39;",
		"-run-product",
		"-client-root &#39;/tmp/client&#39;",
		"-product-storage-state &#39;/tmp/state.json&#39;",
		"-product-riido-api-host &#39;https://development.api.riido.io&#39;",
		"-product-team-id &#39;team-a&#39;",
		"-product-start-client",
		"-product-task-id &#39;task-a&#39;",
		"-product-agent-id-1 &#39;agent-a&#39;",
		"-product-agent-id-2 &#39;agent-b&#39;",
		".local/bin",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("plist missing %q:\n%s", want, got)
		}
	}
}

func TestShellQuoteEscapesSingleQuote(t *testing.T) {
	if got := shellQuote("a'b"); got != `'a'"'"'b'` {
		t.Fatalf("quote=%q", got)
	}
}
