package supervisor

import (
	"context"
	"errors"
	"testing"
)

type fixtureSyncHTTPError struct{}

func (fixtureSyncHTTPError) Error() string       { return "secret-body-and-token" }
func (fixtureSyncHTTPError) HTTPStatusCode() int { return 401 }
func TestRuntimeSyncFailureNeverExposesErrorText(t *testing.T) {
	for _, tc := range []struct {
		err    error
		kind   string
		status int
	}{{context.DeadlineExceeded, "timeout", 0}, {context.Canceled, "cancelled", 0}, {fixtureSyncHTTPError{}, "http", 401}, {errors.New("secret-body-and-token"), "internal", 0}} {
		kind, status := runtimeSyncFailure(tc.err)
		if kind != tc.kind || status != tc.status {
			t.Fatalf("kind=%q status=%d", kind, status)
		}
	}
}
