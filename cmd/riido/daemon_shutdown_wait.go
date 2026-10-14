package main

import "time"

const daemonShutdownPollInterval = 50 * time.Millisecond

func waitDaemonShutdownCondition(timeout time.Duration, done func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if done() {
			return true
		}
		time.Sleep(daemonShutdownPollInterval)
	}
	return false
}
