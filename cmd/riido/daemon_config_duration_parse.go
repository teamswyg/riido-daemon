package main

import (
	"strconv"
	"strings"
	"time"
)

func parseOptionalNonNegativeInt(raw, name string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, daemonWrapf(ErrDaemonConfig, "settings.parse-non-negative-int", err, "%s must be a non-negative integer", name)
	}
	return n, nil
}

func parseOptionalDurationSeconds(raw, name string) (time.Duration, error) {
	n, err := parseOptionalNonNegativeInt(raw, name)
	if err != nil {
		return 0, err
	}
	return time.Duration(n) * time.Second, nil
}

func parseOptionalPositiveDurationSeconds(raw, name string, fallback time.Duration) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0, daemonWrapf(ErrDaemonConfig, "settings.parse-positive-int", err, "%s must be a positive integer", name)
	}
	return time.Duration(n) * time.Second, nil
}
