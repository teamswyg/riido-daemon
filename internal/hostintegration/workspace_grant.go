package hostintegration

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// WorkspaceGrantMethod records the OS/user action that makes a user workspace
// root accessible to Riido.
type WorkspaceGrantMethod string

const (
	WorkspaceGrantDevLocalPath             WorkspaceGrantMethod = "dev-local-path"
	WorkspaceGrantUserSelectedFolder       WorkspaceGrantMethod = "user-selected-folder"
	WorkspaceGrantSecurityScopedBookmark   WorkspaceGrantMethod = "security-scoped-bookmark"
	WorkspaceGrantWindowsFolderPickerGrant WorkspaceGrantMethod = "windows-folder-picker-grant"
)

// WorkspaceGrantRecord is the local-only C11 fact that a specific user
// workspace root may be materialized into C6 workdirs.
type WorkspaceGrantRecord struct {
	WorkspaceID string
	Channel     DistributionChannel
	HostOS      HostOS
	Method      WorkspaceGrantMethod
	RootPath    string
	Label       string
	GrantedBy   string
	GrantedAt   time.Time
	RevokedAt   time.Time
}

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

// Revoked reports whether the grant has been revoked.
func (r WorkspaceGrantRecord) Revoked() bool {
	return !r.RevokedAt.IsZero()
}

// MaterializationAllowed reports whether C6 may materialize this user
// workspace root, given the latest consent view.
func (r WorkspaceGrantRecord) MaterializationAllowed(consent ConsentState) bool {
	return !r.Revoked() && consent.WorkspaceAccessAllowed(r.WorkspaceID)
}

// Validate checks C11 grant rules. It does not inspect security-scoped bookmark
// bytes or Windows shell item tokens; adapters own that OS-specific proof.
func (r WorkspaceGrantRecord) Validate() error {
	var errs []error
	workspaceID := strings.TrimSpace(r.WorkspaceID)
	rootPath := strings.TrimSpace(r.RootPath)
	if workspaceID == "" {
		errs = append(errs, errors.New("workspace id is required"))
	}
	if !r.Channel.Valid() {
		errs = append(errs, fmt.Errorf("unknown distribution channel %q", r.Channel))
	}
	if !r.HostOS.Valid() {
		errs = append(errs, fmt.Errorf("unknown host OS %q", r.HostOS))
	}
	if !r.Method.Valid() {
		errs = append(errs, fmt.Errorf("unknown workspace grant method %q", r.Method))
	}
	if rootPath == "" {
		errs = append(errs, errors.New("workspace root path is required"))
	}
	if r.GrantedAt.IsZero() {
		errs = append(errs, errors.New("granted time is required"))
	}
	if !r.RevokedAt.IsZero() && r.RevokedAt.Before(r.GrantedAt) {
		errs = append(errs, errors.New("revoked time must be after granted time"))
	}
	if r.Channel.Valid() && r.HostOS.Valid() && r.Method.Valid() {
		if err := validateWorkspaceGrantChannelMethod(r.Channel, r.HostOS, r.Method); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Valid reports whether method is one of the SSOT-defined workspace grant
// methods.
func (method WorkspaceGrantMethod) Valid() bool {
	switch method {
	case WorkspaceGrantDevLocalPath,
		WorkspaceGrantUserSelectedFolder,
		WorkspaceGrantSecurityScopedBookmark,
		WorkspaceGrantWindowsFolderPickerGrant:
		return true
	default:
		return false
	}
}

func validateWorkspaceGrantChannelMethod(channel DistributionChannel, hostOS HostOS, method WorkspaceGrantMethod) error {
	switch channel {
	case DistributionChannelDevLocal:
		if method != WorkspaceGrantDevLocalPath && method != WorkspaceGrantUserSelectedFolder {
			return fmt.Errorf("%s workspace grant requires dev-local path or user-selected folder", channel)
		}
	case DistributionChannelDeveloperID:
		if hostOS != HostOSDarwin {
			return errors.New("developer-id workspace grant requires darwin host OS")
		}
		if method != WorkspaceGrantUserSelectedFolder && method != WorkspaceGrantSecurityScopedBookmark {
			return fmt.Errorf("%s workspace grant requires user-selected folder or security-scoped bookmark", channel)
		}
	case DistributionChannelMacAppStore:
		if hostOS != HostOSDarwin {
			return errors.New("mac-app-store workspace grant requires darwin host OS")
		}
		if method != WorkspaceGrantSecurityScopedBookmark {
			return errors.New("mac-app-store workspace grant requires security-scoped bookmark")
		}
	case DistributionChannelMSIXSideload:
		if hostOS != HostOSWindows {
			return errors.New("msix-sideload workspace grant requires windows host OS")
		}
		if method != WorkspaceGrantUserSelectedFolder && method != WorkspaceGrantWindowsFolderPickerGrant {
			return errors.New("msix-sideload workspace grant requires user-selected folder or windows folder picker grant")
		}
	case DistributionChannelMSIXStore:
		if hostOS != HostOSWindows {
			return errors.New("msix-store workspace grant requires windows host OS")
		}
		if method != WorkspaceGrantWindowsFolderPickerGrant {
			return errors.New("msix-store workspace grant requires windows folder picker grant")
		}
	}
	return nil
}
