package detectutil

import (
	"os"
	"testing"
)

func TestEnvListWithLaunchPATHFromMapFallsBackToLaunchPATH(t *testing.T) {
	overrideAugmentedSearchDirs(t, "/launch/bin", "/tool/bin")
	got := EnvListWithLaunchPATHFromMap([]string{"RIIDO_TEST=1"}, nil)
	path, ok := envListValue(got, pathEnvKey())
	want := "/launch/bin" + string(os.PathListSeparator) + "/tool/bin"
	if !ok || path != want {
		t.Fatalf("PATH = %q ok=%v env=%v", path, ok, got)
	}
}

func TestEnvListWithLaunchPATHFromMapSkipsEmptyFallback(t *testing.T) {
	overrideAugmentedSearchDirs(t)
	got := EnvListWithLaunchPATHFromMap([]string{"RIIDO_TEST=1"}, nil)
	if _, ok := envListValue(got, pathEnvKey()); ok {
		t.Fatalf("unexpected PATH in env=%v", got)
	}
}

func TestExtractLoginShellPATHRejectsMalformedMarkers(t *testing.T) {
	if got := extractLoginShellPATH("plain output"); got != "" {
		t.Fatalf("unexpected PATH from plain output: %q", got)
	}
	if got := extractLoginShellPATH(loginPATHMarkerEnd + "/bin" + loginPATHMarkerStart); got != "" {
		t.Fatalf("unexpected PATH from reversed markers: %q", got)
	}
	out := "noise" + loginPATHMarkerStart + "/login/bin" + loginPATHMarkerEnd + "tail"
	if got := extractLoginShellPATH(out); got != "/login/bin" {
		t.Fatalf("PATH = %q", got)
	}
}

func TestCandidateCollectorDropsEmptyAndDuplicatePaths(t *testing.T) {
	var collector = newCandidateCollector()
	collector.add("")
	collector.add("/tmp/tool")
	collector.add("/tmp/../tmp/tool")
	if len(collector.values) != 1 || collector.values[0] != "/tmp/tool" {
		t.Fatalf("candidate values=%v", collector.values)
	}
}
