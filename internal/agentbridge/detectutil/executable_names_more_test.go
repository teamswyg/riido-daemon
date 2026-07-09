package detectutil

import (
	"os"
	"runtime"
	"slices"
	"testing"
)

func TestWindowsExecutableNamesUsePATHEXTEntries(t *testing.T) {
	t.Setenv("PATHEXT", ".EXE"+string(os.PathListSeparator)+" .CMD "+string(os.PathListSeparator))
	got := windowsExecutableNames("riido")
	want := []string{"riido.EXE", "riido.CMD"}
	if !slices.Equal(got, want) {
		t.Fatalf("windows executable names=%v want %v", got, want)
	}
}

func TestExecutableNamesKeepsUnixNameAndExplicitExtension(t *testing.T) {
	if runtime.GOOS != "windows" {
		got := executableNames("riido")
		if len(got) != 1 || got[0] != "riido" {
			t.Fatalf("unix executable names=%v", got)
		}
	}
	got := executableNames("riido.exe")
	if len(got) != 1 || got[0] != "riido.exe" {
		t.Fatalf("explicit extension names=%v", got)
	}
}
