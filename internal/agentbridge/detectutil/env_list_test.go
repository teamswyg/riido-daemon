package detectutil

import "testing"

func TestEnvListWithLaunchPATHFromMapUsesFrozenPath(t *testing.T) {
	got := EnvListWithLaunchPATHFromMap(
		[]string{"RIIDO_TEST=1"},
		map[string]string{pathEnvKey(): "/frozen/bin"},
	)
	path, ok := envListValue(got, pathEnvKey())
	if !ok || path != "/frozen/bin" {
		t.Fatalf("spawn PATH = %q ok=%v, env=%v", path, ok, got)
	}
}

func TestEnvListWithLaunchPATHFromMapPreservesSpawnPath(t *testing.T) {
	got := EnvListWithLaunchPATHFromMap(
		[]string{pathEnvKey() + "=/spawn/bin"},
		map[string]string{pathEnvKey(): "/frozen/bin"},
	)
	path, ok := envListValue(got, pathEnvKey())
	if !ok || path != "/spawn/bin" {
		t.Fatalf("spawn PATH = %q ok=%v, env=%v", path, ok, got)
	}
}

func TestEnvListWithLaunchPATHFromMapPreservesExplicitEmptyPath(t *testing.T) {
	got := EnvListWithLaunchPATHFromMap(nil, map[string]string{pathEnvKey(): ""})
	path, ok := envListValue(got, pathEnvKey())
	if !ok || path != "" {
		t.Fatalf("spawn PATH = %q ok=%v, env=%v", path, ok, got)
	}
}
