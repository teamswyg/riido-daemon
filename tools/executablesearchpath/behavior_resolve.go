package main

import (
	"os"
	"path/filepath"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

func verifyPathOrder() error {
	firstDir, secondDir, err := twoExecutableDirs("riido-path-order-tool")
	if err != nil {
		return err
	}
	defer os.RemoveAll(firstDir)
	defer os.RemoveAll(secondDir)
	return withPATH(firstDir+string(os.PathListSeparator)+secondDir, func() error {
		got := detectutil.ResolveExecutableCandidates("riido-path-order-tool", "")
		if len(got) < 2 {
			return behaviorError("path order did not find both candidates")
		}
		if got[0] != executablePath(firstDir, "riido-path-order-tool") ||
			got[1] != executablePath(secondDir, "riido-path-order-tool") {
			return behaviorError("path order did not preserve process PATH order")
		}
		return nil
	})
}

func twoExecutableDirs(name string) (string, string, error) {
	firstDir, err := tempDir()
	if err != nil {
		return "", "", err
	}
	secondDir, err := tempDir()
	if err != nil {
		_ = os.RemoveAll(firstDir)
		return "", "", err
	}
	if _, err := writeExecutable(firstDir, name); err != nil {
		return "", "", err
	}
	if _, err := writeExecutable(secondDir, name); err != nil {
		return "", "", err
	}
	return firstDir, secondDir, nil
}

func executablePath(dir, name string) string {
	return filepath.Clean(filepath.Join(dir, name))
}
