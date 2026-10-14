package main

import "testing"

func TestTimestampSlug(t *testing.T) {
	got := timestampSlug("2026-06-22T07:32:49Z")
	if got != "20260622T073249Z" {
		t.Fatalf("slug=%q", got)
	}
}

func TestTrimTrailingSlash(t *testing.T) {
	got := trimTrailingSlash("s3://bucket/daily///")
	if got != "s3://bucket/daily" {
		t.Fatalf("prefix=%q", got)
	}
}

func TestOutputPathPreservesAbsolutePath(t *testing.T) {
	got := outputPath("/repo", "/tmp/evidence.json")
	if got != "/tmp/evidence.json" {
		t.Fatalf("path=%q", got)
	}
}
