package main

import "testing"

func TestManualBreakdownRanksDirectories(t *testing.T) {
	docs := []docClass{
		{Path: "docs/20-domain/workspace/a.md", Kind: "manual_registered"},
		{Path: "docs/20-domain/workspace/b.md", Kind: "manual_registered"},
		{Path: "docs/migration/daemon/runtime/a.md", Kind: "manual_registered"},
		{Path: "docs/README.md", Kind: "generated"},
	}
	got := manualTopDirs(docs, 2)
	if len(got) != 2 {
		t.Fatalf("hotspot len = %d", len(got))
	}
	if got[0].Path != "docs/20-domain/workspace" || got[0].Count != 2 {
		t.Fatalf("first hotspot = %#v", got[0])
	}
}

func TestManualSamplesLimitPerGroup(t *testing.T) {
	docs := []docClass{
		{Path: "a.md", Kind: "manual_registered", Group: "g", Reason: "r"},
		{Path: "b.md", Kind: "manual_registered", Group: "g", Reason: "r"},
		{Path: "c.md", Kind: "manual_registered", Group: "h", Reason: "s"},
	}
	got := manualSamples(docs, 1)
	if len(got) != 2 {
		t.Fatalf("samples len = %d", len(got))
	}
	if got[0].Path != "a.md" || got[1].Path != "c.md" {
		t.Fatalf("samples = %#v", got)
	}
}
