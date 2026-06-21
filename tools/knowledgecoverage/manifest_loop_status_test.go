package main

import (
	"path/filepath"
	"testing"
)

func TestManifestLoopStatusAcceptsLoopSource(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "source.riido.json", manifestLoopFixture())
	writeFixture(t, root, "target.riido.json", `{"loop_source":"source.riido.json"}`)
	got := manifestLoopStatus(root, filepath.Join(root, "target.riido.json"))
	if got != "delegated" {
		t.Fatalf("status = %q", got)
	}
}

func TestScanManifestLoopsCountsMissingDebt(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "looped.riido.json", manifestLoopFixture())
	writeFixture(t, root, "missing.riido.json", `{}`)
	got, err := scanManifestLoops(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Direct != 1 || got.Missing != 1 {
		t.Fatalf("loops = %#v", got)
	}
}

func manifestLoopFixture() string {
	return `{"loop":{"observation":"o","hypothesis":"h","execute":"x","evaluate":"e","retrospective":"r"}}`
}
