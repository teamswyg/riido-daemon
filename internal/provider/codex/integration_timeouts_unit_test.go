package codex

import "testing"

func TestIntegrationTimeoutsLeaveCleanupWindow(t *testing.T) {
	if codexIntegrationContextTimeout <= codexIntegrationHardTimeout {
		t.Fatal("context timeout must exceed session hard timeout")
	}
}
