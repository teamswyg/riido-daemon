package main

import "testing"

func TestDaemonForegroundStartsAndExposesStatus(t *testing.T) {
	run := startForegroundDaemonForStatus(t)
	status, out := readDaemonStatus(t, run.socket)

	assertDaemonRuntimes(t, status, out)
	assertDaemonStatusFields(t, status, run.socket)
	assertDaemonMetrics(t, status)
	assertForegroundDaemonExits(t, run.cancel, run.errCh)
}
