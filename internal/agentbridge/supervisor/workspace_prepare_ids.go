package supervisor

import (
	"errors"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

type workspaceIDs struct {
	workspaceID   string
	logicalTaskID string
	runID         string
}

func workspacePrepareIDs(req *bridge.TaskRequest) (workspaceIDs, error) {
	workspaceID := firstMetadata(req.Metadata, MetadataWorkspaceID, MetadataWorkspace)
	if workspaceID == "" {
		return workspaceIDs{}, errors.New("supervisor: workspace_id metadata is required when Workdir is configured")
	}
	logicalTaskID := firstMetadata(req.Metadata, controlplane.MetadataTaskID)
	if logicalTaskID == "" {
		logicalTaskID = req.ID
	}
	runID := firstMetadata(req.Metadata, MetadataRunID)
	if runID == "" {
		runID = req.ID
	}
	return workspaceIDs{workspaceID: workspaceID, logicalTaskID: logicalTaskID, runID: runID}, nil
}

func workspacePrepareCapability(status runtimeactor.Status, req *bridge.TaskRequest) (runtimeactor.Capability, error) {
	capView, ok := findCapability(status.Capabilities, string(req.Provider))
	if !ok {
		return runtimeactor.Capability{}, fmt.Errorf("supervisor: capability for provider %q disappeared before workspace prepare", req.Provider)
	}
	return capView, nil
}
