package main

import "testing"

func TestScanManifestLoopsDelegatesPointerFileInputs(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "source.riido.json", manifestLoopFixture())
	writeFixture(t, root, "owner.riido.json", `{
		"loop_source":"source.riido.json",
		"package_roles_file":"fragments/package-roles.riido.json"
	}`)
	writeFixture(t, root, "fragments/package-roles.riido.json", `[{"role":"r"}]`)
	got, err := scanManifestLoops(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Direct != 1 || got.Delegated != 2 || got.Missing != 0 {
		t.Fatalf("loops = %#v", got)
	}
}

func TestScanManifestLoopsDelegatesRootRelativePointerFileInputs(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "docs/owner.riido.json", `{
		"loop":{"observation":"o","hypothesis":"h","execute":"x","evaluate":"e","retrospective":"r"},
		"provider_validation_manifest":"docs/provider-validation.riido.json"
	}`)
	writeFixture(t, root, "docs/provider-validation.riido.json", `{}`)
	got, err := scanManifestLoops(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Direct != 1 || got.Delegated != 1 || got.Missing != 0 {
		t.Fatalf("loops = %#v", got)
	}
}

func TestScanManifestLoopsDelegatesPointerArrayInputs(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "source.riido.json", manifestLoopFixture())
	writeFixture(t, root, "owner.riido.json", `{
		"loop_source":"source.riido.json",
		"provider_files":["providers/claude.riido.json"]
	}`)
	writeFixture(t, root, "providers/claude.riido.json", `{"provider":"claude"}`)
	got, err := scanManifestLoops(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Direct != 1 || got.Delegated != 2 || got.Missing != 0 {
		t.Fatalf("loops = %#v", got)
	}
}

func TestScanManifestLoopsDoesNotDelegatePointerFileFromMissingOwner(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "owner.riido.json", `{
		"package_roles_file":"fragments/package-roles.riido.json"
	}`)
	writeFixture(t, root, "fragments/package-roles.riido.json", `[{"role":"r"}]`)
	got, err := scanManifestLoops(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Delegated != 0 || got.Missing != 2 {
		t.Fatalf("loops = %#v", got)
	}
}
