package hostintegration

import (
	"reflect"
	"testing"
)

func TestExternalToolRecordServerFacingStatusStripsPrivatePath(t *testing.T) {
	record := validExternalToolRecord()

	status, err := record.ServerFacingStatus(DistributionChannelMacAppStore, "1.2.3")
	if err != nil {
		t.Fatalf("server status failed: %v", err)
	}
	if status.DistributionChannel != DistributionChannelMacAppStore ||
		status.AppVersion != "1.2.3" ||
		status.ProviderKind != record.Provider ||
		!status.ProviderAvailable ||
		status.ProviderLoginStatus != ToolLoginLoggedIn {
		t.Fatalf("unexpected status: %+v", status)
	}

	statusType := reflect.TypeOf(status)
	for _, forbidden := range []string{"ExecutablePath", "WorkspaceRootPath", "Token", "APIKey"} {
		if _, ok := statusType.FieldByName(forbidden); ok {
			t.Fatalf("server-facing status leaked forbidden field %s", forbidden)
		}
	}
}
