package hostintegration

import "time"

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

// Revoked reports whether the grant has been revoked.
func (r WorkspaceGrantRecord) Revoked() bool {
	return !r.RevokedAt.IsZero()
}

// MaterializationAllowed reports whether C6 may materialize this user
// workspace root, given the latest consent view.
func (r WorkspaceGrantRecord) MaterializationAllowed(consent ConsentState) bool {
	return !r.Revoked() && consent.WorkspaceAccessAllowed(r.WorkspaceID)
}
