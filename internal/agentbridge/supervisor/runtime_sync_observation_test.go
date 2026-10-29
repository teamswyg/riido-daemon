package supervisor

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"time"
)

func (fixtureSyncHTTPError) RequestCorrelationID() string { return "0123456789abcdef0123456789abcdef" }
func TestRuntimeSyncFailureIncludesLastSuccessWithoutSensitiveText(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	defer log.SetOutput(previous)
	a := &Actor{}
	a.runtimeSyncSuccess.Store("runtime", time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC))
	a.logRuntimeSyncFailure("heartbeat_sync_failed", "runtime", fixtureSyncHTTPError{})
	got := output.String()
	for _, want := range []string{"http_status=401", "request_id=\"0123456789abcdef0123456789abcdef\"", "last_sync_success_at=\"2026-10-29T00:00:00Z\""} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %s: %s", want, got)
		}
	}
	if strings.Contains(got, "secret-body-and-token") {
		t.Fatal("sensitive error leaked")
	}
}
