package main

import "testing"

func TestScanManifestLoopsDelegatesFragmentInputs(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "source.riido.json", manifestLoopFixture())
	writeFixture(t, root, "owner.riido.json", `{
		"loop_source":"source.riido.json",
		"fragments":{"overview":"fragments/overview.riido.json"}
	}`)
	writeFixture(t, root, "fragments/overview.riido.json", `{"title":"overview"}`)
	got, err := scanManifestLoops(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Direct != 1 || got.Delegated != 2 || got.Missing != 0 {
		t.Fatalf("loops = %#v", got)
	}
}

func TestScanManifestLoopsDoesNotDelegateFragmentFromMissingOwner(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "owner.riido.json", `{
		"fragments":{"overview":"fragments/overview.riido.json"}
	}`)
	writeFixture(t, root, "fragments/overview.riido.json", `{"title":"overview"}`)
	got, err := scanManifestLoops(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Delegated != 0 || got.Missing != 2 {
		t.Fatalf("loops = %#v", got)
	}
}
