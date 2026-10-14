package hostintegration_test

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
)

func TestWorkspaceGrantBDDStoreChannelRequiresOSGrant(t *testing.T) {
	record := workspaceGrantRecord()
	record.Channel = hostintegration.DistributionChannelMacAppStore
	record.HostOS = hostintegration.HostOSDarwin
	record.Method = hostintegration.WorkspaceGrantUserSelectedFolder

	err := record.Validate()

	requireWorkspaceGrantValidationError(t, err, "security-scoped bookmark")
}

func TestWorkspaceGrantBDDMSIXStoreRequiresFolderPickerGrant(t *testing.T) {
	record := workspaceGrantRecord()
	record.Channel = hostintegration.DistributionChannelMSIXStore
	record.HostOS = hostintegration.HostOSWindows
	record.Method = hostintegration.WorkspaceGrantUserSelectedFolder
	record.RootPath = `C:\Users\tester\repo`

	err := record.Validate()

	requireWorkspaceGrantValidationError(t, err, "windows folder picker grant")
}
