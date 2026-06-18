package ingest

import (
	"errors"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
)

// Ingestor owns CanonicalEvent envelope completion and validation.
type Ingestor struct {
	cfg Config
}

func New(cfg Config) (*Ingestor, error) {
	if cfg.Sink == nil {
		return nil, errors.New("ingest: Sink is required")
	}
	if strings.TrimSpace(cfg.RiidoDaemonVersion) == "" {
		return nil, errors.New("ingest: RiidoDaemonVersion is required")
	}
	if strings.TrimSpace(cfg.PolicyBundleVersion) == "" {
		return nil, errors.New("ingest: PolicyBundleVersion is required")
	}
	if cfg.ActorKind == "" {
		cfg.ActorKind = ir.ActorDaemon
	}
	if cfg.ActorKind != ir.ActorSystem && strings.TrimSpace(cfg.ActorID) == "" {
		return nil, errors.New("ingest: ActorID is required unless ActorKind=system")
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.NewEventID == nil {
		cfg.NewEventID = NewUUID7EventID
	}
	return &Ingestor{cfg: cfg}, nil
}
