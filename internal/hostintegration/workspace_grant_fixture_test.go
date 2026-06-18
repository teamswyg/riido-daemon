package hostintegration_test

import (
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
)

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

func requireWorkspaceGrantValidationError(
	t *testing.T,
	err error,
	wantSubstring string,
) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected validation error containing %q", wantSubstring)
	}
	if !strings.Contains(err.Error(), wantSubstring) {
		t.Fatalf("error = %q, want %s", err, wantSubstring)
	}
}
