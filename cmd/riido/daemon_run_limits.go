package main

import (
	"strings"
	"time"
)

func loadRunLimits(getenv func(string) string) (time.Duration, time.Duration, error) {
	hard, err := runLimit(getenv(envRunHardTimeout), envRunHardTimeout, 30*time.Minute)
	if err != nil {
		return 0, 0, err
	}
	idle, err := runLimit(getenv(envRunSemanticIdle), envRunSemanticIdle, 3*time.Minute)
	return hard, idle, err
}

func runLimit(raw, name string, fallback time.Duration) (time.Duration, error) {
	if strings.TrimSpace(raw) == "" {
		return fallback, nil
	}
	return parseOptionalDurationSeconds(raw, name)
}
