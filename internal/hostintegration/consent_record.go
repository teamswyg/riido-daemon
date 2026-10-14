package hostintegration

import (
	"time"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

// ConsentRecord is one immutable user-intent fact. Provider and WorkspaceID
// are mutually exclusive subjects depending on ConsentKind.
type ConsentRecord struct {
	Kind        ConsentKind
	Decision    ConsentDecision
	Provider    capability.ProviderKind
	WorkspaceID string
	Actor       string
	Reason      string
	RecordedAt  time.Time
}
