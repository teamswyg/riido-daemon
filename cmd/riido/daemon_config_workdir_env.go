package main

import (
	"strings"
	"time"
)

type daemonWorkspaceEnv struct {
	root                 string
	count                int
	runtimeMaxConcurrent int
	retention            time.Duration
	cleanupEvery         time.Duration
}

func loadDaemonWorkspaceEnv(getenv func(string) string, userHome func() (string, error)) (daemonWorkspaceEnv, error) {
	count, err := parseOptionalNonNegativeInt(getenv(envWorkspaceCount), envWorkspaceCount)
	if err != nil {
		return daemonWorkspaceEnv{}, err
	}
	maxConcurrent, err := loadRuntimeMaxConcurrent(getenv)
	if err != nil {
		return daemonWorkspaceEnv{}, err
	}
	root, err := loadDaemonWorkdirRoot(getenv, userHome)
	if err != nil {
		return daemonWorkspaceEnv{}, err
	}
	retention, cleanupEvery, err := loadDaemonWorkdirCleanup(getenv)
	if err != nil {
		return daemonWorkspaceEnv{}, err
	}
	return daemonWorkspaceEnv{
		root:                 root,
		count:                count,
		runtimeMaxConcurrent: maxConcurrent,
		retention:            retention,
		cleanupEvery:         cleanupEvery,
	}, nil
}

func loadRuntimeMaxConcurrent(getenv func(string) string) (int, error) {
	maxConcurrent, err := parseOptionalNonNegativeInt(getenv(envRuntimeMaxConcurrent), envRuntimeMaxConcurrent)
	if err != nil {
		return 0, err
	}
	if maxConcurrent == 0 {
		return 4, nil
	}
	return maxConcurrent, nil
}

func loadDaemonWorkdirRoot(getenv func(string) string, userHome func() (string, error)) (string, error) {
	root := strings.TrimSpace(getenv(envWorkdirRoot))
	if root != "" {
		return root, nil
	}
	return defaultAgentDaemonWorkdirRoot(userHome)
}
