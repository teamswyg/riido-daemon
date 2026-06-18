package detectutil

import "testing"

func TestLoginShellPATHDirsCached(t *testing.T) {
	calls := 0
	restore := readLoginShellPATH
	readLoginShellPATH = func() string {
		calls++
		return "/tmp/cached-bin"
	}
	resetLoginShellCacheForTest()
	t.Cleanup(func() {
		readLoginShellPATH = restore
		resetLoginShellCacheForTest()
	})

	_ = loginShellPATHDirs()
	_ = loginShellPATHDirs()
	if calls != 1 {
		t.Fatalf("login-shell PATH should be read once, got %d reads", calls)
	}
}
