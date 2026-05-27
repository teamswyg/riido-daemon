package hostintegration_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
)

func TestWorkspaceGrantBDDStoreChannelRequiresOSGrant(t *testing.T) {
	// Given a Mac App Store distribution where arbitrary home scanning is not allowed.
	record := workspaceGrantRecord()
	record.Channel = hostintegration.DistributionChannelMacAppStore
	record.HostOS = hostintegration.HostOSDarwin
	record.Method = hostintegration.WorkspaceGrantUserSelectedFolder

	// When the grant is validated with only a plain folder path.
	err := record.Validate()

	// Then the domain rejects it and requires a security-scoped bookmark.
	if err == nil {
		t.Fatal("expected mac app store plain folder grant to fail")
	}
	if !strings.Contains(err.Error(), "security-scoped bookmark") {
		t.Fatalf("error = %q, want security-scoped bookmark", err)
	}
}

func TestWorkspaceGrantBDDMSIXStoreRequiresFolderPickerGrant(t *testing.T) {
	// Given a Microsoft Store MSIX distribution.
	record := workspaceGrantRecord()
	record.Channel = hostintegration.DistributionChannelMSIXStore
	record.HostOS = hostintegration.HostOSWindows
	record.Method = hostintegration.WorkspaceGrantUserSelectedFolder
	record.RootPath = `C:\Users\tester\repo`

	// When a plain path grant is submitted.
	err := record.Validate()

	// Then the domain rejects it in favor of a Windows folder picker grant.
	if err == nil {
		t.Fatal("expected msix store plain folder grant to fail")
	}
	if !strings.Contains(err.Error(), "windows folder picker grant") {
		t.Fatalf("error = %q, want windows folder picker grant", err)
	}
}

func TestWorkspaceGrantBDDConsentAndGrantAreBothRequired(t *testing.T) {
	// Given a valid store workspace grant.
	record := workspaceGrantRecord()
	record.Channel = hostintegration.DistributionChannelMacAppStore
	record.HostOS = hostintegration.HostOSDarwin
	record.Method = hostintegration.WorkspaceGrantSecurityScopedBookmark
	store, err := hostintegration.NewWorkspaceGrantStore(record)
	if err != nil {
		t.Fatal(err)
	}

	// When the user has not granted workspace-access consent.
	grant, ok := store.ActiveGrant("workspace-1")
	if !ok {
		t.Fatal("expected active grant")
	}
	withoutConsent := hostintegration.ConsentState{WorkspaceAccess: map[string]bool{}}

	// Then C6 materialization is still blocked.
	if grant.MaterializationAllowed(withoutConsent) {
		t.Fatal("grant without workspace-access consent should not materialize")
	}

	// And when matching consent exists, materialization becomes allowed.
	withConsent := hostintegration.ConsentState{WorkspaceAccess: map[string]bool{"workspace-1": true}}
	if !grant.MaterializationAllowed(withConsent) {
		t.Fatal("grant with workspace-access consent should materialize")
	}
}

func TestWorkspaceGrantBDDRevokedGrantBlocksMaterialization(t *testing.T) {
	// Given a previously accepted grant that has been revoked.
	record := workspaceGrantRecord()
	record.RevokedAt = record.GrantedAt.Add(time.Minute)
	store, err := hostintegration.NewWorkspaceGrantStore(record)
	if err != nil {
		t.Fatal(err)
	}

	// When the same workspace is requested.
	_, ok := store.ActiveGrant("workspace-1")

	// Then no active grant is returned.
	if ok {
		t.Fatal("revoked grant should not be active")
	}
}

func TestWorkspaceGrantStoreRecordsAreDeterministic(t *testing.T) {
	first := workspaceGrantRecord()
	first.WorkspaceID = "workspace-b"
	second := workspaceGrantRecord()
	second.WorkspaceID = "workspace-a"

	store, err := hostintegration.NewWorkspaceGrantStore(first, second)
	if err != nil {
		t.Fatal(err)
	}

	gotRecords := store.Records()
	got := []string{gotRecords[0].WorkspaceID, gotRecords[1].WorkspaceID}
	want := []string{"workspace-a", "workspace-b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("records order = %v, want %v", got, want)
	}
}

func workspaceGrantRecord() hostintegration.WorkspaceGrantRecord {
	return hostintegration.WorkspaceGrantRecord{
		WorkspaceID: "workspace-1",
		Channel:     hostintegration.DistributionChannelDevLocal,
		HostOS:      hostintegration.HostOSDarwin,
		Method:      hostintegration.WorkspaceGrantDevLocalPath,
		RootPath:    "/Users/tester/repo",
		Label:       "tester repo",
		GrantedBy:   "user:tester",
		GrantedAt:   time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC),
	}
}
