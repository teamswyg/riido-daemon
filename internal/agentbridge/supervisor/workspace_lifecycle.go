package supervisor

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/execution"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func (a *Actor) prepareWorkspace(ctx context.Context, status runtimeactor.Status, req *bridge.TaskRequest) (*preparedWorkspace, error) {
	if req.ReferenceFolders != nil && !req.ReferenceFolders.Matches(req.Metadata["execution_device_id"], status.RuntimeID) {
		return nil, errors.New("reference folders belong to a different device or runtime; select the folders again")
	}
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
	req.HostExecutionBinding = a.nativeSessionExecutionBinding(req)
	req.HostExecutionBindingSelected = true
	capView, err := workspacePrepareCapability(status, req)
	if err != nil {
		return nil, err
	}
	var retained workdir.Workspace
	resumed, err := a.restoreNativeSessionWorkspace(req, capView, &retained)
	if err != nil {
		return nil, err
	}
	ws, err := a.cfg.Workdir.Prepare(workdir.TaskID{Workspace: ids.workspaceID, Task: ids.logicalTaskID, Run: ids.runID})
	if err != nil {
		return nil, err
	}
	if resumed {
		ws.Workdir = retained.Workdir
	}
	events, err := a.newWorkspaceEventContext(ws, status.RuntimeID, req, ids.logicalTaskID, ids.runID, capView)
	if err != nil {
		return nil, err
	}
	a.appendWorkspaceEvent(ctx, req.ID, events, ir.EventWorkdirCreated, "", map[string]any{
		"workdirPath": ws.Workdir,
		"taskID":      ids.logicalTaskID,
	})
	if !resumed {
		if err := materializeAssignmentWorktree(ctx, ws.Workdir, req.Worktree); err != nil {
			return nil, err
		}
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
	prepared := &preparedWorkspace{workspace: &ws, events: events, canResume: capView.SupportsResume}
	if selected := req.HostExecutionBinding; selected != nil {
		binding := *selected
		if resumed {
			prepared.execution, err = execution.Reopen(binding, ws.Workdir)
		} else {
			prepared.execution, err = execution.Prepare(ctx, binding, ws.Workdir)
		}
		if err != nil {
			if resumed {
				return nil, fmt.Errorf("%w: retained repository could not be reopened: %w", errNativeSessionBinding, err)
			}
			return nil, err
		}
		req.Cwd = prepared.execution.Directory
		appendRepositoryInstructions(req, binding)
	}
	return prepared, nil
}

func appendRepositoryInstructions(req *bridge.TaskRequest, binding execution.Binding) {
	req.Prompt += fmt.Sprintf("\n\nThe authorized source repository is at %s. Perform the requested work there and validate the changes.", req.Cwd)
	if binding.HasChecks() {
		req.Prompt += " A successful reply alone is not completion; the host will require changed files and passing checks."
	} else {
		req.Prompt += " Host verification is not configured. Run appropriate checks and report what was actually verified; do not claim host-verified completion."
	}
}
