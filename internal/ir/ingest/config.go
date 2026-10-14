package ingest

import (
	"context"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
)

const EventSchemaVersionV1 = 1

// Sink is the persistence port behind EventIngestor.
type Sink interface {
	AppendEvents(context.Context, []ir.CanonicalEvent) error
}

// Config fixes server-decided event envelope fields.
type Config struct {
	Sink                Sink
	RiidoDaemonVersion  string
	PolicyBundleVersion string
	ActorKind           ir.ActorKind
	ActorID             string
	Now                 func() time.Time
	NewEventID          func(time.Time) (string, error)
}
