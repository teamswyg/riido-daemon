package hostintegration

import (
	"time"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

func validExternalToolRecord() ExternalToolRecord {
	return ExternalToolRecord{
		Provider:            "codex",
		ExecutablePath:      "/usr/local/bin/codex",
		Provenance:          ToolProvenanceUserSelected,
		DetectedVersion:     "0.1.0",
		LoginStatus:         ToolLoginLoggedIn,
		CompatibilityStatus: capability.CompatSupported,
		LastVerifiedAt:      time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC),
	}
}
