package userauth

import (
	"errors"
	"net/http"
	"testing"
)

func TestExistingDeviceLinkConflictRequiresExactCreationResponse(t *testing.T) {
	for _, tc := range []struct {
		name       string
		status     int
		recovery   bool
		body, want string
	}{
		{"create", 409, false, `{"error":"installation_already_linked","error_description":"untrusted secret"}`, "LINK_ALREADY_LINKED"},
		{"recovery", 409, true, `{"error":"installation_already_linked","error_description":"ignored"}`, "LINK_RECOVERY_REJECTED"},
		{"server failure", 503, false, `{"error":"installation_already_linked","error_description":"ignored"}`, "LINK_HTTP_ERROR"},
		{"unknown field", 409, false, `{"error":"installation_already_linked","error_description":"ignored","device_id":"private"}`, "LINK_REQUEST_REJECTED"},
		{"unknown policy", 503, false, `{"error":"installation_policy_unknown","error_description":"ignored"}`, "LINK_POLICY_UNKNOWN"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got *DeviceLinkError
			if err := deviceLinkHTTPError(tc.status, []byte(tc.body), tc.recovery); !errors.As(err, &got) || got.Code != tc.want || got.StatusCode != tc.status {
				t.Fatal("wrong fixed error classification", err)
			}
			if got.Code == "LINK_ALREADY_LINKED" && (got.Retryable || got.StatusCode != http.StatusConflict) {
				t.Fatal("existing link must not be retried")
			}
		})
	}
}
