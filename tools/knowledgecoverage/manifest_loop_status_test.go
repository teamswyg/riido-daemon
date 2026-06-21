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

func TestScanManifestLoopsDelegatesEvidenceFileFragments(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "source.riido.json", manifestLoopFixture())
	writeFixture(t, root, "owner.riido.json", `{
		"loop_source":"source.riido.json",
		"evidence_files":{"local":["fragments/local.riido.json"]}
	}`)
	writeFixture(t, root, "fragments/local.riido.json", `[{"risk":"r"}]`)
	got, err := scanManifestLoops(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Direct != 1 || got.Delegated != 2 || got.Missing != 0 {
		t.Fatalf("loops = %#v", got)
	}
}

func TestScanManifestLoopsDoesNotDelegateFromMissingOwner(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "owner.riido.json", `{
		"evidence_files":{"local":["fragments/local.riido.json"]}
	}`)
	writeFixture(t, root, "fragments/local.riido.json", `[{"risk":"r"}]`)
	got, err := scanManifestLoops(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Delegated != 0 || got.Missing != 2 {
		t.Fatalf("loops = %#v", got)
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
