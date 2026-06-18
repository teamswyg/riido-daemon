package main

import "time"

func loadDaemonWorkdirCleanup(getenv func(string) string) (time.Duration, time.Duration, error) {
	retention, err := parseOptionalDurationSeconds(getenv(envWorkdirRetentionSeconds), envWorkdirRetentionSeconds)
	if err != nil {
		return 0, 0, err
	}
	cleanupEvery, err := parseOptionalDurationSeconds(getenv(envWorkdirCleanupIntervalSeconds), envWorkdirCleanupIntervalSeconds)
	if err != nil {
		return 0, 0, err
	}
	if cleanupEvery > 0 && retention <= 0 {
		return 0, 0, daemonErrorf(ErrDaemonConfig, "settings.validate-workdir-retention", "%s requires %s", envWorkdirCleanupIntervalSeconds, envWorkdirRetentionSeconds)
	}
	if retention > 0 && cleanupEvery <= 0 {
		cleanupEvery = time.Hour
	}
	return retention, cleanupEvery, nil
}
