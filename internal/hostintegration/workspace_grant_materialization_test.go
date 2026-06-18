package hostintegration_test

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
)

func TestWorkspaceGrantBDDConsentAndGrantAreBothRequired(t *testing.T) {
	record := workspaceGrantRecord()
	record.Channel = hostintegration.DistributionChannelMacAppStore
	record.HostOS = hostintegration.HostOSDarwin
	record.Method = hostintegration.WorkspaceGrantSecurityScopedBookmark

	store, err := hostintegration.NewWorkspaceGrantStore(record)
	if err != nil {
		t.Fatal(err)
	}

	grant, ok := store.ActiveGrant("workspace-1")
	if !ok {
		t.Fatal("expected active grant")
	}

	withoutConsent := hostintegration.ConsentState{WorkspaceAccess: map[string]bool{}}
	if grant.MaterializationAllowed(withoutConsent) {
		t.Fatal("grant without workspace-access consent should not materialize")
	}

	withConsent := hostintegration.ConsentState{
		WorkspaceAccess: map[string]bool{"workspace-1": true},
	}
	if !grant.MaterializationAllowed(withConsent) {
		t.Fatal("grant with workspace-access consent should materialize")
	}
}

func TestWorkspaceGrantBDDRevokedGrantBlocksMaterialization(t *testing.T) {
	record := workspaceGrantRecord()
	record.RevokedAt = record.GrantedAt.Add(time.Minute)

	store, err := hostintegration.NewWorkspaceGrantStore(record)
	if err != nil {
		t.Fatal(err)
	}

	_, ok := store.ActiveGrant("workspace-1")
	if ok {
		t.Fatal("revoked grant should not be active")
	}
}
