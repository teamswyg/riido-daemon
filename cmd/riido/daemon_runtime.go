package main

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/taskdbplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolpolicy"
	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/workdir"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func codexRuntimeModels(userHome func() (string, error)) []runtimeactor.RuntimeModel {
	modelID := codexConfiguredModelID(userHome)
	if modelID == "" {
		return nil
	}
	return []runtimeactor.RuntimeModel{{
		ModelID:   modelID,
		Label:     modelID,
		IsDefault: true,
	}}
}

func codexConfiguredModelID(userHome func() (string, error)) string {
	if userHome == nil {
		return ""
	}
	home, err := userHome()
	if err != nil || strings.TrimSpace(home) == "" {
		return ""
	}
	body, err := os.ReadFile(filepath.Join(home, ".codex", "config.toml"))
	if err != nil {
		return ""
	}
	scanner := bufio.NewScanner(strings.NewReader(string(body)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, rawValue, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) != "model" {
			continue
		}
		value := strings.TrimSpace(rawValue)
		if unquoted, err := strconv.Unquote(value); err == nil {
			return strings.TrimSpace(unquoted)
		}
		if commentAt := strings.Index(value, "#"); commentAt >= 0 {
			value = strings.TrimSpace(value[:commentAt])
			if unquoted, err := strconv.Unquote(value); err == nil {
				return strings.TrimSpace(unquoted)
			}
		}
		return strings.TrimSpace(value)
	}
	return ""
}

func daemonToolAutoApprover(settings daemonSettings) agentbridge.AutoApprover {
	return toolpolicy.PolicyAutoApprover(settings.PolicyBundleDoc, policy.TrustTierHost)
}

func daemonToolStartGate(settings daemonSettings) agentbridge.ToolStartGate {
	return toolpolicy.PolicyToolStartGate(settings.PolicyBundleDoc, policy.TrustTierHost)
}

func stopRuntimeActors(ctx lifecycle.Context, runtimes []*runtimeactor.Actor, log logging.Logger) {
	for _, rt := range runtimes {
		if err := rt.StopLifecycle(ctx); err != nil {
			log.Printf("runtimeactor stop error level=%s: %v", ctx.ShutdownLevel(), err)
		}
	}
}

func providerRuntimeID(daemonID, provider string) string {
	if provider == "" {
		return daemonID
	}
	return daemonID + ":" + provider
}

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

func startWorkdirCleanupLoop(ctx lifecycle.Context, cleaner workdir.Cleaner, settings daemonSettings, log logging.Logger) func() {
	if settings.WorkdirRetention <= 0 {
		return func() {}
	}
	if settings.WorkdirCleanupEvery <= 0 {
		settings.WorkdirCleanupEvery = time.Hour
	}
	cleanupCtx, cancel := lifecycle.WithCancel(ctx)
	runCleanup := func() {
		cutoff := time.Now().UTC().Add(-settings.WorkdirRetention)
		result, err := cleaner.CleanupArchivedBefore(cleanupCtx.Context(), workdir.CleanupRequest{ArchivedBefore: cutoff})
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("workdir cleanup error: %v", err)
			}
			return
		}
		if len(result.Removed) > 0 {
			log.Printf("workdir cleanup removed=%d scanned=%d retention=%s", len(result.Removed), result.ScannedArchiveRecords, settings.WorkdirRetention)
		}
	}
	runCleanup()
	go func() {
		ticker := time.NewTicker(settings.WorkdirCleanupEvery)
		defer ticker.Stop()
		for {
			select {
			case <-cleanupCtx.Done():
				return
			case <-ticker.C:
				runCleanup()
			}
		}
	}()
	return cancel
}

// builtinDaemonAdapters returns the four shipped agent adapters. The
// daemon's RuntimeActor takes ownership of their Detect lifecycle.
func builtinDaemonAdapters() []agentbridge.Adapter {
	return []agentbridge.Adapter{
		bridgeClaudeAdapter{},
		bridgeCodexAdapter{},
		bridgeOpenClawAdapter{},
		bridgeCursorAdapter{},
	}
}
