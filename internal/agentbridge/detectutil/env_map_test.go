package detectutil

import "testing"

func TestEnvMapWithLaunchPATHAddsFrozenPath(t *testing.T) {
	binDir := t.TempDir()
	overrideAugmentedSearchDirs(t, binDir)

	got := EnvMapWithLaunchPATH(map[string]string{"RIIDO_TEST": "1"})
	if got["RIIDO_TEST"] != "1" {
		t.Fatalf("existing env value missing: %+v", got)
	}
	if path := EnvMapPATHValue(got); path != binDir {
		t.Fatalf("PATH = %q, want %q", path, binDir)
	}
}

func TestEnvMapWithLaunchPATHPreservesExplicitPath(t *testing.T) {
	got := EnvMapWithLaunchPATH(map[string]string{pathEnvKey(): "/custom/bin"})
	if path := EnvMapPATHValue(got); path != "/custom/bin" {
		t.Fatalf("PATH = %q, want explicit value", path)
	}
}
