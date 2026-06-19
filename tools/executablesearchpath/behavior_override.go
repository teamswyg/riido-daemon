package main

import (
	"os"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

func verifyOverrideOnly() error {
	overrideDir, err := tempDir()
	if err != nil {
		return err
	}
	pathDir, err := tempDir()
	if err != nil {
		_ = os.RemoveAll(overrideDir)
		return err
	}
	defer os.RemoveAll(overrideDir)
	defer os.RemoveAll(pathDir)
	override, err := writeExecutable(overrideDir, "riido-override-tool")
	if err != nil {
		return err
	}
	if _, err := writeExecutable(pathDir, "riido-override-tool"); err != nil {
		return err
	}
	return withPATH(pathDir, func() error {
		got := detectutil.ResolveExecutableCandidates("riido-override-tool", override)
		if len(got) != 1 || got[0] != override {
			return behaviorError("override was not the only candidate")
		}
		return nil
	})
}

func verifyOverrideFailClosed() error {
	pathDir, err := tempDir()
	if err != nil {
		return err
	}
	defer os.RemoveAll(pathDir)
	if _, err := writeExecutable(pathDir, "riido-missing-override-tool"); err != nil {
		return err
	}
	return withPATH(pathDir, func() error {
		got := detectutil.ResolveExecutableCandidates("riido-missing-override-tool", "/definitely/not/real")
		if len(got) != 0 {
			return behaviorError("missing override fell back to PATH")
		}
		return nil
	})
}
