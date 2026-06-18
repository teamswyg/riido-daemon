package supervisor

import (
	"context"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func (a *Actor) prepareWorkspace(ctx context.Context, status runtimeactor.Status, req *bridge.TaskRequest) (*preparedWorkspace, error) {
	if a.cfg.Workdir == nil {
		return nil, nil
	}
	if req.Metadata == nil {
		req.Metadata = map[string]string{}
	}
	ids, err := workspacePrepareIDs(req)
	if err != nil {
		return nil, err
	}
	capView, err := workspacePrepareCapability(status, req)
	if err != nil {
		return nil, err
	}
	ws, err := a.cfg.Workdir.Prepare(workdir.TaskID{Workspace: ids.workspaceID, Task: ids.logicalTaskID, Run: ids.runID})
	if err != nil {
		return nil, err
	}
	events, err := a.newWorkspaceEventContext(ws, status.RuntimeID, req, ids.logicalTaskID, ids.runID, capView)
	if err != nil {
		return nil, err
	}
	a.appendWorkspaceEvent(ctx, req.ID, events, ir.EventWorkdirCreated, "", map[string]any{
		"workdirPath": ws.Workdir,
		"taskID":      ids.logicalTaskID,
	})
	if err := materializeAssignmentWorktree(ctx, ws.Workdir, req.Worktree); err != nil {
		return nil, err
	}
	native, err := a.resolveNativeWorkspaceConfig(req)
	if err != nil {
		return nil, err
	}
	if err := a.injectWorkspaceRuntimeConfig(ws, req, capView.ProtocolKind, native); err != nil {
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
	applyPreparedWorkspaceMetadata(req, ws, ids.runID, native.resolved, nativeConfigVersion)
	if events != nil {
		events.nativeConfigVersion = nativeConfigVersion
	}
	a.appendWorkspaceEvent(ctx, req.ID, events, ir.EventNativeConfigInjected, nativeConfigVersion, map[string]any{
		"files":               native.resolved.GeneratedFiles(),
		"nativeConfigVersion": nativeConfigVersion,
	})
	return &preparedWorkspace{workspace: &ws, events: events}, nil
}
