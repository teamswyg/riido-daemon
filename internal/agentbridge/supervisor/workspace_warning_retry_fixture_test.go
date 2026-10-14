package supervisor

import (
	"context"
	"errors"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/ir/ingest"
)

func failingWorkspaceEvents(t *testing.T) *workspaceEventContext {
	t.Helper()
	return &workspaceEventContext{
		taskID:     "logical-1",
		runID:      "run-1",
		runtimeID:  "runtime-1",
		capability: failingWorkspaceCapability(),
		ingestor:   failingWorkspaceIngestor(t, ir.ActorDaemon),
	}
}

func failingWorkspaceCapability() runtimeactor.Capability {
	return runtimeactor.Capability{
		Provider:              "codex",
		ProtocolKind:          "codex",
		Version:               "1.0.0",
		CapabilityFingerprint: "fp-1",
	}
}

func failingWorkspaceIngestor(t *testing.T, actorKind ir.ActorKind) *ingest.Ingestor {
	t.Helper()
	ingestor, err := ingest.New(ingest.Config{
		Sink:                failingIngestSink{},
		RiidoDaemonVersion:  "riido-daemon.test.v1",
		PolicyBundleVersion: "policy-bundle.test.v1",
		ActorKind:           actorKind,
		ActorID:             "daemon-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	return ingestor
}

type failingIngestSink struct{}

func (failingIngestSink) AppendEvents(context.Context, []ir.CanonicalEvent) error {
	return errors.New("append rejected")
}
