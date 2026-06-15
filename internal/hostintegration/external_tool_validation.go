package hostintegration

import (
	"sort"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

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
