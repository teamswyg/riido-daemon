package detectutil

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestOSEnvReadsProcessEnvironment(t *testing.T) {
	t.Setenv("RIIDO_DETECTUTIL_TEST", "present")
	if got := OSEnv("RIIDO_DETECTUTIL_TEST"); got != "present" {
		t.Fatalf("OSEnv() = %q", got)
	}
}

func TestEnvListWithLaunchPATHUsesPreferredValue(t *testing.T) {
	got := EnvListWithLaunchPATH([]string{"RIIDO_TEST=1"}, "/preferred/bin")
	path, ok := envListValue(got, pathEnvKey())
	if !ok || path != "/preferred/bin" {
		t.Fatalf("PATH = %q ok=%v env=%v", path, ok, got)
	}
	got[0] = "RIIDO_TEST=mutated"
	if gotAgain := EnvListWithLaunchPATH([]string{"RIIDO_TEST=1"}, "/preferred/bin"); gotAgain[0] != "RIIDO_TEST=1" {
		t.Fatalf("input env should be cloned, got %v", gotAgain)
	}
}

func TestEnvListWithLaunchPATHPreservesExistingPath(t *testing.T) {
	got := EnvListWithLaunchPATH([]string{pathEnvKey() + "=/existing/bin"}, "/preferred/bin")
	path, ok := envListValue(got, pathEnvKey())
	if !ok || path != "/existing/bin" {
		t.Fatalf("PATH = %q ok=%v env=%v", path, ok, got)
	}
}

func TestFallbackExecutableNames(t *testing.T) {
	if got := fallbackExecutableNames("riido", nil); len(got) != 1 || got[0] != "riido" {
		t.Fatalf("fallback name = %v", got)
	}
	values := []string{"riido.EXE"}
	if got := fallbackExecutableNames("riido", values); !slices.Equal(got, values) {
		t.Fatalf("fallback should preserve values, got %v", got)
	}
}

func TestWindowsHomeDirsContainProviderBins(t *testing.T) {
	home := filepath.Join("C:", "Users", "tester")
	got := windowsHomeDirs(home)
	for _, want := range []string{
		filepath.Join(home, ".cursor", "bin"),
		filepath.Join(home, ".claude", "bin"),
		filepath.Join(home, ".cargo", "bin"),
	} {
		if !slices.Contains(got, want) {
			t.Fatalf("windowsHomeDirs missing %s in %v", want, got)
		}
	}
}
