package detectutil

import (
	"os"
	"testing"
)

func TestLaunchPATHUsesAugmentedSearchDirs(t *testing.T) {
	firstDir := t.TempDir()
	secondDir := t.TempDir()
	overrideAugmentedSearchDirs(t, firstDir, secondDir)

	got := LaunchPATH()
	want := firstDir + string(os.PathListSeparator) + secondDir
	if got != want {
		t.Fatalf("LaunchPATH = %q, want %q", got, want)
	}
}
