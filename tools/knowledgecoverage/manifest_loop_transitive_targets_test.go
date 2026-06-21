package main

import "testing"

func TestScanManifestLoopsDelegatesTransitivePointerInputs(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "source.riido.json", manifestLoopFixture())
	writeFixture(t, root, "integration.riido.json", `{
		"loop_source":"source.riido.json",
		"provider_validation_manifest":"docs/provider-validation.riido.json"
	}`)
	writeFixture(t, root, "docs/provider-validation.riido.json", `{
		"provider_files":["provider-validation/claude.riido.json"]
	}`)
	writeFixture(t, root, "docs/provider-validation/claude.riido.json", `{"provider":"claude"}`)
	got, err := scanManifestLoops(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Direct != 1 || got.Delegated != 3 || got.Missing != 0 {
		t.Fatalf("loops = %#v", got)
	}
}
