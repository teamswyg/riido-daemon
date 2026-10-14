package hostintegration

import (
	"errors"
	"sort"
)

// WorkspaceGrantStore is a pure in-memory view. Durable storage and OS grant
// resolution live in adapters; C6 consumes only accepted grant records.
type WorkspaceGrantStore struct {
	records map[string]WorkspaceGrantRecord
}

// NewWorkspaceGrantStore validates and stores grants in deterministic domain
// shape.
func NewWorkspaceGrantStore(records ...WorkspaceGrantRecord) (*WorkspaceGrantStore, error) {
	store := &WorkspaceGrantStore{records: make(map[string]WorkspaceGrantRecord)}
	for _, record := range records {
		if err := store.Put(record); err != nil {
			return nil, err
		}
	}
	return store, nil
}

// Put stores or replaces the current grant for a workspace id.
func (s *WorkspaceGrantStore) Put(record WorkspaceGrantRecord) error {
	if s == nil {
		return errors.New("workspace grant store is nil")
	}
	if err := record.Validate(); err != nil {
		return err
	}
	if s.records == nil {
		s.records = make(map[string]WorkspaceGrantRecord)
	}
	s.records[record.WorkspaceID] = record
	return nil
}

// Lookup returns the current grant for a workspace id.
func (s *WorkspaceGrantStore) Lookup(workspaceID string) (WorkspaceGrantRecord, bool) {
	if s == nil {
		return WorkspaceGrantRecord{}, false
	}
	record, ok := s.records[workspaceID]
	return record, ok
}

// ActiveGrant returns the current, non-revoked grant for a workspace id.
func (s *WorkspaceGrantStore) ActiveGrant(workspaceID string) (WorkspaceGrantRecord, bool) {
	record, ok := s.Lookup(workspaceID)
	if !ok || record.Revoked() {
		return WorkspaceGrantRecord{}, false
	}
	return record, true
}

// Records returns all current grant records sorted by workspace id.
func (s *WorkspaceGrantStore) Records() []WorkspaceGrantRecord {
	if s == nil || len(s.records) == 0 {
		return nil
	}
	records := make([]WorkspaceGrantRecord, 0, len(s.records))
	for _, record := range s.records {
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].WorkspaceID < records[j].WorkspaceID
	})
	return records
}
