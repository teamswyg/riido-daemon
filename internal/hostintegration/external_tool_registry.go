package hostintegration

import (
	"errors"
	"sort"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

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
