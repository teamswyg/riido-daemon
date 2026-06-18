package hostintegration

import "testing"

func TestExternalToolRecordStoreAutoDetectedRequiresConfirmation(t *testing.T) {
	record := validExternalToolRecord()
	record.Provenance = ToolProvenanceAutoDetected

	if !record.RequiresExecutionConfirmation(DistributionChannelMacAppStore) {
		t.Fatal("auto-detected CLI should require confirmation in mac app store channel")
	}
	if !record.RequiresExecutionConfirmation(DistributionChannelMSIXStore) {
		t.Fatal("auto-detected CLI should require confirmation in msix store channel")
	}
	if record.RequiresExecutionConfirmation(DistributionChannelDeveloperID) {
		t.Fatal("developer-id channel should not use the store confirmation rule")
	}
}
