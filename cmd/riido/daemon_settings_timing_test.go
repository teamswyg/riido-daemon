package main

import (
	"testing"
	"time"
)

func TestLoadDaemonSettingsReadsQueueAndTiming(t *testing.T) {
	settings := loadDaemonSettingsForTest(t, fullDaemonSettingsEnv())
	if settings.TaskQueueDir != "/tmp/riido-queue" || settings.TaskReportDir != "/tmp/riido-reports" {
		t.Fatalf("task queue/report dirs mismatch: %+v", settings)
	}
	if settings.WorkdirRetention != 24*time.Hour || settings.WorkdirCleanupEvery != 5*time.Minute {
		t.Fatalf("workdir cleanup settings mismatch: %+v", settings)
	}
	if settings.PollEvery != 7*time.Second ||
		settings.IdlePollEvery != 21*time.Second ||
		settings.HeartbeatEvery != 30*time.Second {
		t.Fatalf("poll/heartbeat settings mismatch: %+v", settings)
	}
}
