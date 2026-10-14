package supervisor

import (
	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/ir/ingest"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func (a *Actor) newWorkspaceEventContext(ws workdir.Workspace, statusRuntimeID string, req *bridge.TaskRequest, logicalTaskID, runID string, capView runtimeactor.Capability) (*workspaceEventContext, error) {
	sink, err := workdir.NewRunEventSink(ws)
	if err != nil {
		return nil, err
	}
	ingestor, err := ingest.New(ingest.Config{
		Sink:                sink,
		RiidoDaemonVersion:  a.cfg.RiidoDaemonVersion,
		PolicyBundleVersion: a.cfg.PolicyBundleVersion,
		ActorKind:           ir.ActorDaemon,
		ActorID:             a.cfg.DaemonID,
	})
	if err != nil {
		return nil, err
	}
	agentIngestor, err := ingest.New(ingest.Config{
		Sink:                sink,
		RiidoDaemonVersion:  a.cfg.RiidoDaemonVersion,
		PolicyBundleVersion: a.cfg.PolicyBundleVersion,
		ActorKind:           ir.ActorAgent,
		ActorID:             runID,
	})
	if err != nil {
		return nil, err
	}
	return &workspaceEventContext{
		taskID:        logicalTaskID,
		runID:         runID,
		runtimeID:     statusRuntimeID,
		capability:    capView,
		ingestor:      ingestor,
		agentIngestor: agentIngestor,
	}, nil
}
