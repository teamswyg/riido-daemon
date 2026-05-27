package hostintegration

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

// DistributionChannel is the package artifact identity that constrains which
// host surfaces may be used.
type DistributionChannel string

const (
	DistributionChannelDeveloperID  DistributionChannel = "developer-id"
	DistributionChannelMacAppStore  DistributionChannel = "mac-app-store"
	DistributionChannelMSIXSideload DistributionChannel = "msix-sideload"
	DistributionChannelMSIXStore    DistributionChannel = "msix-store"
	DistributionChannelDevLocal     DistributionChannel = "dev-local"
)

// ToolProvenance records why Riido trusts an executable path enough to show it
// as a provider candidate.
type ToolProvenance string

const (
	ToolProvenanceUserSelected ToolProvenance = "user-selected"
	ToolProvenanceEnvOverride  ToolProvenance = "env-override"
	ToolProvenanceAutoDetected ToolProvenance = "auto-detected"
)

// ToolLoginStatus is intentionally not a failure enum. LoginRequired means the
// provider is real but not currently routable.
type ToolLoginStatus string

const (
	ToolLoginUnknown  ToolLoginStatus = "unknown"
	ToolLoginLoggedIn ToolLoginStatus = "logged-in"
	ToolLoginRequired ToolLoginStatus = "login-required"
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

// ServerFacingToolStatus is the privacy-filtered subset of a registration row
// that may cross into C10. It intentionally has no executable path, workspace
// path, token, or provider secret fields.
type ServerFacingToolStatus struct {
	DistributionChannel DistributionChannel     `json:"distribution_channel"`
	AppVersion          string                  `json:"app_version,omitempty"`
	ProviderKind        capability.ProviderKind `json:"provider_kind"`
	ProviderAvailable   bool                    `json:"provider_available"`
	ProviderLoginStatus ToolLoginStatus         `json:"provider_login_status"`
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

// ServerFacingStatus returns the C10-safe projection of the local registration
// row. Do not add path-like fields here; distribution-host-integration.md §7 is
// the SSOT for this boundary.
func (r ExternalToolRecord) ServerFacingStatus(channel DistributionChannel, appVersion string) (ServerFacingToolStatus, error) {
	if !channel.Valid() {
		return ServerFacingToolStatus{}, fmt.Errorf("unknown distribution channel %q", channel)
	}
	if err := r.Validate(); err != nil {
		return ServerFacingToolStatus{}, err
	}
	return ServerFacingToolStatus{
		DistributionChannel: channel,
		AppVersion:          strings.TrimSpace(appVersion),
		ProviderKind:        r.Provider,
		ProviderAvailable:   r.ProviderAvailable(),
		ProviderLoginStatus: r.LoginStatus,
	}, nil
}

// ExternalToolRegistry is a pure in-memory view of registered provider CLIs.
// Persistence and OS discovery live outside this package.
type ExternalToolRegistry struct {
	records map[capability.ProviderKind]ExternalToolRecord
}

// NewExternalToolRegistry creates a registry and applies the same provenance
// precedence rules as Register.
func NewExternalToolRegistry(records ...ExternalToolRecord) (*ExternalToolRegistry, error) {
	registry := &ExternalToolRegistry{records: make(map[capability.ProviderKind]ExternalToolRecord)}
	for _, record := range records {
		if _, _, err := registry.Register(record); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

// Register validates a record and stores it when its provenance is at least as
// authoritative as the current row for that provider. It returns the effective
// row plus whether the supplied record became effective.
func (r *ExternalToolRegistry) Register(record ExternalToolRecord) (ExternalToolRecord, bool, error) {
	if r == nil {
		return ExternalToolRecord{}, false, errors.New("registry is nil")
	}
	if err := record.Validate(); err != nil {
		return ExternalToolRecord{}, false, err
	}
	if r.records == nil {
		r.records = make(map[capability.ProviderKind]ExternalToolRecord)
	}
	current, ok := r.records[record.Provider]
	if ok && provenanceRank(record.Provenance) < provenanceRank(current.Provenance) {
		return current, false, nil
	}
	r.records[record.Provider] = record
	return record, true, nil
}

// Lookup returns the effective row for a provider.
func (r *ExternalToolRegistry) Lookup(provider capability.ProviderKind) (ExternalToolRecord, bool) {
	if r == nil {
		return ExternalToolRecord{}, false
	}
	record, ok := r.records[provider]
	return record, ok
}

// Records returns a deterministic snapshot sorted by ProviderKind.
func (r *ExternalToolRegistry) Records() []ExternalToolRecord {
	if r == nil || len(r.records) == 0 {
		return nil
	}
	records := make([]ExternalToolRecord, 0, len(r.records))
	for _, record := range r.records {
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].Provider < records[j].Provider
	})
	return records
}

// Valid reports whether channel is one of the SSOT-defined distribution
// channels.
func (c DistributionChannel) Valid() bool {
	switch c {
	case DistributionChannelDeveloperID,
		DistributionChannelMacAppStore,
		DistributionChannelMSIXSideload,
		DistributionChannelMSIXStore,
		DistributionChannelDevLocal:
		return true
	default:
		return false
	}
}

// StoreManaged reports whether the channel is subject to app store review
// constraints.
func (c DistributionChannel) StoreManaged() bool {
	return c == DistributionChannelMacAppStore || c == DistributionChannelMSIXStore
}

// Valid reports whether provenance is one of the SSOT-defined provenance values.
func (p ToolProvenance) Valid() bool {
	switch p {
	case ToolProvenanceUserSelected, ToolProvenanceEnvOverride, ToolProvenanceAutoDetected:
		return true
	default:
		return false
	}
}

// Valid reports whether status is one of the SSOT-defined login statuses.
func (s ToolLoginStatus) Valid() bool {
	switch s {
	case ToolLoginUnknown, ToolLoginLoggedIn, ToolLoginRequired:
		return true
	default:
		return false
	}
}

func provenanceRank(provenance ToolProvenance) int {
	switch provenance {
	case ToolProvenanceUserSelected:
		return 3
	case ToolProvenanceEnvOverride:
		return 2
	case ToolProvenanceAutoDetected:
		return 1
	default:
		return 0
	}
}

func validCompatibilityStatus(status capability.CompatibilityStatus) bool {
	switch status {
	case capability.CompatSupported,
		capability.CompatDegraded,
		capability.CompatExperimental,
		capability.CompatBlocked:
		return true
	default:
		return false
	}
}
