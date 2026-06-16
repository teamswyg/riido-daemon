package hostintegration

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

// ExternalToolRecord is the C11 registration record for one external provider
// CLI path. The executable path is local-only and must never be sent to C10.
type ExternalToolRecord struct {
	Provider            capability.ProviderKind
	ExecutablePath      string
	Provenance          ToolProvenance
	DetectedVersion     string
	LoginStatus         ToolLoginStatus
	CompatibilityStatus capability.CompatibilityStatus
	LastVerifiedAt      time.Time
}

// Validate checks the domain-level invariants for a registry row. It does not
// check whether the path exists or whether the provider can execute; those are
// adapter/probe responsibilities.
func (r ExternalToolRecord) Validate() error {
	var errs []error
	if strings.TrimSpace(string(r.Provider)) == "" {
		errs = append(errs, errors.New("provider is required"))
	}
	if strings.TrimSpace(r.ExecutablePath) == "" {
		errs = append(errs, errors.New("executable path is required"))
	}
	if !r.Provenance.Valid() {
		errs = append(errs, fmt.Errorf("unknown provenance %q", r.Provenance))
	}
	if !r.LoginStatus.Valid() {
		errs = append(errs, fmt.Errorf("unknown login status %q", r.LoginStatus))
	}
	if !validCompatibilityStatus(r.CompatibilityStatus) {
		errs = append(errs, fmt.Errorf("unknown compatibility status %q", r.CompatibilityStatus))
	}
	if r.LastVerifiedAt.IsZero() {
		errs = append(errs, errors.New("last verified time is required"))
	}
	return errors.Join(errs...)
}

// ProviderAvailable returns whether this provider can be routed to from a
// store/server status perspective. Login-required is actionable status, not an
// execution candidate.
func (r ExternalToolRecord) ProviderAvailable() bool {
	return r.LoginStatus == ToolLoginLoggedIn &&
		r.CompatibilityStatus != capability.CompatBlocked
}

// RequiresExecutionConfirmation reports whether a user confirmation is needed
// before executing this CLI under the given distribution channel.
func (r ExternalToolRecord) RequiresExecutionConfirmation(channel DistributionChannel) bool {
	return r.Provenance == ToolProvenanceAutoDetected && channel.StoreManaged()
}
