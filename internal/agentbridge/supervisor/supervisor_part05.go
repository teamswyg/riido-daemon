package supervisor

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func (a *Actor) prepareWorkspace(ctx context.Context, status runtimeactor.Status, req *bridge.TaskRequest) (*preparedWorkspace, error) {
	if a.cfg.Workdir == nil {
		return nil, nil
	}
	if req.Metadata == nil {
		req.Metadata = map[string]string{}
	}
	workspaceID := firstMetadata(req.Metadata, MetadataWorkspaceID, MetadataWorkspace)
	if workspaceID == "" {
		return nil, errors.New("supervisor: workspace_id metadata is required when Workdir is configured")
	}
	logicalTaskID := firstMetadata(req.Metadata, controlplane.MetadataTaskID)
	if logicalTaskID == "" {
		logicalTaskID = req.ID
	}
	runID := firstMetadata(req.Metadata, MetadataRunID)
	if runID == "" {
		runID = req.ID
	}
	capView, ok := findCapability(status.Capabilities, string(req.Provider))
	if !ok {
		return nil, fmt.Errorf("supervisor: capability for provider %q disappeared before workspace prepare", req.Provider)
	}
	ws, err := a.cfg.Workdir.Prepare(workdir.TaskID{Workspace: workspaceID, Task: logicalTaskID, Run: runID})
	if err != nil {
		return nil, err
	}
	events, err := a.newWorkspaceEventContext(ws, status.RuntimeID, req, logicalTaskID, runID, capView)
	if err != nil {
		return nil, err
	}
	a.appendWorkspaceEvent(ctx, req.ID, events, ir.EventWorkdirCreated, "", map[string]any{
		"workdirPath": ws.Workdir,
		"taskID":      logicalTaskID,
	})
	nativePlan := workdir.ProviderConfigPlan(string(req.Provider))
	nativeHookMode := a.nativeHookMode(nativePlan)
	nativeConfigHomeMode := a.nativeConfigHomeMode(nativePlan)
	resolvedNativePlan, err := workdir.ResolveProviderConfigPlanWithOptions(string(req.Provider), workdir.ProviderConfigPlanOptions{
		NativeHookMode:       nativeHookMode,
		NativeConfigHomeMode: nativeConfigHomeMode,
	})
	if err != nil {
		return nil, err
	}
	if err := a.cfg.Workdir.InjectRuntimeConfig(ws, workdir.RuntimeConfig{
		Provider:                   string(req.Provider),
		ProtocolKind:               capView.ProtocolKind,
		TelemetryContractPlacement: req.Metadata[agentbridge.MetadataTelemetryContract],
		NativeHookMode:             nativeHookMode,
		NativeConfigHomeMode:       nativeConfigHomeMode,
		Identity:                   runtimeIdentity(req.Metadata),
		HardRules:                  runtimeHardRules(req.Metadata),
		Workflow:                   req.Metadata[MetadataWorkflow],
	}); err != nil {
		return nil, err
	}
	nativeConfigVersion, err := workdir.ComputeNativeConfigVersion(ws, workdir.NativeConfigVersionInput{
		PolicyBundleVersion: a.cfg.PolicyBundleVersion,
		ProviderKind:        string(req.Provider),
		ProtocolKind:        capView.ProtocolKind,
	})
	if err != nil {
		return nil, err
	}
	req.Cwd = ws.Workdir
	req.Metadata[MetadataRunID] = runID
	req.Metadata[MetadataWorkdirRoot] = ws.Root
	req.Metadata[MetadataWorkdir] = ws.Workdir
	req.Metadata[MetadataOutputDir] = ws.Output
	req.Metadata[MetadataLogsDir] = ws.Logs
	req.Metadata[MetadataArtifactsDir] = ws.Artifacts
	req.Metadata[MetadataNativeConfig] = ws.NativeConfig
	if resolvedNativePlan.ConfigHomeDir != "" {
		req.Metadata[MetadataNativeConfigHome] = filepath.Join(ws.Workdir, filepath.FromSlash(resolvedNativePlan.ConfigHomeDir))
	} else {
		delete(req.Metadata, MetadataNativeConfigHome)
	}
	req.Metadata[MetadataIRDir] = ws.IR
	req.Metadata[MetadataNativeConfigVersion] = nativeConfigVersion
	if events != nil {
		events.nativeConfigVersion = nativeConfigVersion
	}
	a.appendWorkspaceEvent(ctx, req.ID, events, ir.EventNativeConfigInjected, nativeConfigVersion, map[string]any{
		"files":               resolvedNativePlan.GeneratedFiles(),
		"nativeConfigVersion": nativeConfigVersion,
	})
	return &preparedWorkspace{workspace: &ws, events: events}, nil
}

func (a *Actor) nativeHookMode(plan workdir.ProviderNativeConfigPlan) string {
	switch plan.HookMode {
	case workdir.NativeConfigHookModeClaudeCommandHooks:
		decision := policy.EvaluateNativeConfigHookWithBundle(a.cfg.PolicyBundle, policy.NativeConfigHookInput{
			TrustTier: a.cfg.RuntimeTrustTier,
			Surface:   policy.NativeConfigHookClaudeCommandAudit,
		})
		if decision.Allowed {
			return plan.HookMode
		}
		return workdir.NativeConfigHookModeInstructionOnly
	default:
		return plan.HookMode
	}
}

func (a *Actor) nativeConfigHomeMode(plan workdir.ProviderNativeConfigPlan) string {
	if plan.ProviderKind == "codex" && plan.ConfigHomeDir == ".codex" {
		decision := policy.EvaluateNativeConfigFileWithBundle(a.cfg.PolicyBundle, policy.NativeConfigFileInput{
			TrustTier: a.cfg.RuntimeTrustTier,
			Surface:   policy.NativeConfigFileCodexTaskScopedHome,
		})
		if decision.Allowed {
			return ""
		}
		return workdir.NativeConfigHomeModeDisabled
	}
	return ""
}

func (a *Actor) recordTerminalResult(ctx context.Context, running *runningTask, res agentbridge.Result) agentbridge.Result {
	if running == nil {
		return res
	}
	if res.Workdir == "" && running.workspace != nil {
		res.Workdir = running.workspace.Workdir
	}
	a.appendTerminalResultEvent(ctx, running.taskID, running.events, res)
	a.archiveTerminalWorkspace(ctx, running.taskID, running.workspace, running.events, res)
	return res
}

func (a *Actor) archiveTerminalWorkspace(ctx context.Context, taskID string, ws *workdir.Workspace, events *workspaceEventContext, res agentbridge.Result) {
	if ws == nil || a.cfg.Workdir == nil {
		return
	}
	archiver, ok := a.cfg.Workdir.(workdir.Archiver)
	if !ok {
		return
	}
	record, err := archiver.Archive(*ws, workdir.ArchiveRequest{
		ResultStatus: string(res.Status),
		ArchivedAt:   res.FinishedAt,
	})
	if err == nil {
		a.appendWorkspaceEvent(ctx, taskID, events, ir.EventWorkdirArchived, eventNativeConfigVersion(events), map[string]any{
			"workdirPath": record.WorkdirPath,
			"archiveURI":  record.ArchiveURI,
		})
		return
	}
	_ = a.cfg.Reporter.ReportEvent(ctx, taskID, agentbridge.Event{
		Kind: agentbridge.EventWarning,
		Text: "workspace archive failed",
		Err:  err.Error(),
	})
}
