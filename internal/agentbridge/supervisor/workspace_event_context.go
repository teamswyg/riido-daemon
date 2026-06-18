package supervisor

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/ir/ingest"
)

type workspaceEventContext struct {
	taskID              string
	runID               string
	runtimeID           string
	capability          runtimeactor.Capability
	nativeConfigVersion string
	ingestor            *ingest.Ingestor
	agentIngestor       *ingest.Ingestor
}
