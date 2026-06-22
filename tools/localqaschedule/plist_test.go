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
		"-run-product",
		"-client-root &#39;/tmp/client&#39;",
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

func testConfig() config {
	repo, s3 := ".", "s3://bucket/daily"
	product, label, plist := "/tmp/product.json", "io.test", ""
	clientRoot, baseURL, workspace := "/tmp/client", "http://localhost:3000", "W1"
	hour, minute := 9, 5
	install, runAtLoad, runProduct := false, false, true
	return config{
		repo:             &repo,
		s3Prefix:         &s3,
		productEvidence:  &product,
		clientRoot:       &clientRoot,
		productBaseURL:   &baseURL,
		productWorkspace: &workspace,
		runProduct:       &runProduct,
		label:            &label,
		plistPath:        &plist,
		hour:             &hour,
		minute:           &minute,
		install:          &install,
		runAtLoad:        &runAtLoad,
	}
}
