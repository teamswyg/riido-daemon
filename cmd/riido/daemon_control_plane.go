package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/taskdbplane"
)

func buildDaemonControlPlane(settings daemonSettings, startedAt time.Time) (controlplane.TaskSourcePort, controlplane.TaskReporterPort, string, error) {
	if settings.SaaSURL != "" {
		plane, err := saasplane.New(saasplane.Config{
			BaseURL:      settings.SaaSURL,
			DaemonID:     settings.DaemonID,
			DeviceID:     settings.DeviceID,
			DeviceSecret: settings.DeviceSecret,
			Profile:      settings.Profile,
			AppVersion:   settings.DaemonVersion,
			PID:          os.Getpid(),
			StartedAt:    startedAt.UTC(),
		})
		if err != nil {
			return nil, nil, "", daemonWrapf(ErrDaemonControlPlane, "control-plane.saas", err, "controlplane: saas source")
		}
		return plane, plane, "saas", nil
	}
	if settings.TaskDBSourcePath != "" {
		if settings.TaskQueueDir != "" {
			return nil, nil, "", daemonErrorf(ErrDaemonConfig, "control-plane.validate-config", "%s cannot be combined with %s", envTaskDBSourcePath, envTaskQueueDir)
		}
		if settings.TaskReportDir != "" {
			return nil, nil, "", daemonErrorf(ErrDaemonConfig, "control-plane.validate-config", "%s cannot be combined with %s", envTaskDBSourcePath, envTaskReportDir)
		}
		plane, err := taskdbplane.New(settings.TaskDBSourcePath)
		if err != nil {
			return nil, nil, "", daemonWrapf(ErrDaemonControlPlane, "control-plane.taskdb", err, "controlplane: task DB source")
		}
		return plane, plane, "taskdb", nil
	}
	if settings.TaskQueueDir == "" {
		if settings.TaskReportDir != "" {
			return nil, nil, "", daemonErrorf(ErrDaemonConfig, "control-plane.validate-config", "%s requires %s", envTaskReportDir, envTaskQueueDir)
		}
		return controlplane.NewMemorySource(), controlplane.NewMemoryReporter(), "memory", nil
	}
	source, err := controlplane.NewFileQueueSource(settings.TaskQueueDir)
	if err != nil {
		return nil, nil, "", daemonWrapf(ErrDaemonControlPlane, "control-plane.file-source", err, "controlplane: file queue source")
	}
	reportDir := settings.TaskReportDir
	if reportDir == "" {
		reportDir = filepath.Join(settings.TaskQueueDir, "reports")
	}
	reporter, err := controlplane.NewFileReporter(reportDir)
	if err != nil {
		return nil, nil, "", daemonWrapf(ErrDaemonControlPlane, "control-plane.file-reporter", err, "controlplane: file reporter")
	}
	return source, reporter, "file", nil
}
