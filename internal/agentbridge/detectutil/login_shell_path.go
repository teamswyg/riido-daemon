package detectutil

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	readLoginShellPATH = defaultLoginShellPATH
	userHomeDir        = os.UserHomeDir
)

var (
	loginShellMu       sync.Mutex
	loginShellResolved bool
	loginShellCache    []string
)

func loginShellPATHDirs() []string {
	loginShellMu.Lock()
	defer loginShellMu.Unlock()
	if !loginShellResolved {
		loginShellCache = filepath.SplitList(readLoginShellPATH())
		loginShellResolved = true
	}
	return loginShellCache
}

func resetLoginShellCacheForTest() {
	loginShellMu.Lock()
	loginShellResolved = false
	loginShellCache = nil
	loginShellMu.Unlock()
}
