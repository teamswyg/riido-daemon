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
	hour, minute := 9, 5
	install, runAtLoad := false, false
	return config{&repo, &s3, &product, &label, &plist, &hour, &minute, &install, &runAtLoad}
}
