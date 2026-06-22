package main

import (
	"os"
	"path/filepath"
)

func resolvePaths(cfg config) (schedulePaths, error) {
	repo, err := filepath.Abs(*cfg.repo)
	if err != nil {
		return schedulePaths{}, err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return schedulePaths{}, err
	}
	plist := *cfg.plistPath
	if plist == "" {
		plist = filepath.Join(home, "Library", "LaunchAgents", *cfg.label+".plist")
	}
	logDir := filepath.Join(repo, ".riido-local", "logs")
	return schedulePaths{
		repo:      repo,
		plist:     plist,
		stdout:    filepath.Join(logDir, "local-qa-launchd.out.log"),
		stderr:    filepath.Join(logDir, "local-qa-launchd.err.log"),
		launchctl: "/bin/launchctl",
	}, nil
}
