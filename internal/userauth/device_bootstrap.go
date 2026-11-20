package userauth

import (
	"context"
	"strconv"
	"strings"
)

const deviceBootstrapPath = transferPath + "/device-bootstraps"

type BootstrapIdentity struct {
	Resource       string `json:"resource"`
	ClientID       string `json:"client_id"`
	InstallationID string `json:"installation_id"`
	KeyFingerprint string `json:"key_fingerprint"`
	OwnerID        string `json:"owner_principal_id"`
}

type BootstrapRequest struct {
	BootstrapIdentity
	DeviceID          string `json:"expected_device_id"`
	RequestID         string `json:"request_id"`
	RecoverySecret    string `json:"recovery_secret"`
	NextHash          string `json:"next_secret_hash"`
	ExpectedExpiresAt string `json:"expected_expires_at,omitempty"`
}

func (BootstrapRequest) String() string     { return "[redacted device bootstrap request]" }
func (r BootstrapRequest) GoString() string { return r.String() }

func ParseBootstrapRequest(raw []byte, restart bool) (BootstrapRequest, error) {
	var input BootstrapRequest
	keys := []string{"resource", "client_id", "installation_id", "key_fingerprint", "owner_principal_id", "request_id", "expected_device_id", "recovery_secret", "next_secret_hash"}
	if restart {
		keys = append(keys, "expected_expires_at")
	}
	fields, err := deviceLinkObject(raw, keys...)
	if err != nil || decodeDeviceLink(fields, &input) != nil || !deviceLinkRandom(input.RequestID) || !deviceLinkRandom(input.RecoverySecret) || !deviceLinkDigest(input.NextHash) || !deviceLinkDigest(input.KeyFingerprint) || !deviceLinkID(input.InstallationID) || !deviceLinkID(input.OwnerID) || input.ClientID != "riido-cli" || len(input.DeviceID) != 36 || !strings.HasPrefix(input.DeviceID, "dev_") || strings.Trim(input.DeviceID[4:], "0123456789abcdef") != "" {
		return BootstrapRequest{}, linkFailure("BOOTSTRAP_INVALID_REQUEST")
	}
	if restart {
		n, e := strconv.ParseInt(input.ExpectedExpiresAt, 10, 64)
		if e != nil || n <= 0 || strconv.FormatInt(n, 10) != input.ExpectedExpiresAt {
			return BootstrapRequest{}, linkFailure("BOOTSTRAP_INVALID_REQUEST")
		}
	}
	return input, nil
}

func (m *Manager) BootstrapIdentity(ctx context.Context, session string) (BootstrapIdentity, error) {
	raw, status, err := m.installationSessionPost(ctx, session, deviceBootstrapPath+"/identity", map[string]string{})
	if err != nil {
		return BootstrapIdentity{}, err
	}
	if status != 200 {
		return BootstrapIdentity{}, deviceLinkHTTPError(status, raw, false)
	}
	var id BootstrapIdentity
	fields, err := deviceLinkObject(raw, "resource", "client_id", "installation_id", "key_fingerprint", "owner_principal_id")
	if err != nil || decodeDeviceLink(fields, &id) != nil || id.Resource != m.client.config.Resource || id.ClientID != "riido-cli" || !deviceLinkDigest(id.KeyFingerprint) || !deviceLinkID(id.OwnerID) || !deviceLinkID(id.InstallationID) {
		return BootstrapIdentity{}, linkFailure("BOOTSTRAP_INVALID_RESPONSE")
	}
	return id, nil
}

func (m *Manager) RequestBootstrap(ctx context.Context, session string, input BootstrapRequest) (DeviceLinkApproval, error) {
	path := deviceBootstrapPath
	payload := map[string]string{"request_id": input.RequestID, "expected_device_id": input.DeviceID, "recovery_secret": input.RecoverySecret, "next_secret_hash": input.NextHash, "installation_id": input.InstallationID, "key_fingerprint": input.KeyFingerprint}
	if input.ExpectedExpiresAt != "" {
		path += "/restart"
		payload["expected_expires_at"] = input.ExpectedExpiresAt
	}
	raw, status, err := m.installationSessionPost(ctx, session, path, payload)
	if err != nil {
		return DeviceLinkApproval{}, err
	}
	if status != 200 {
		return DeviceLinkApproval{}, deviceLinkHTTPError(status, raw, false)
	}
	var result DeviceLinkApproval
	fields, err := deviceLinkObject(raw, "request_id", "state", "expires_at")
	if err != nil || decodeDeviceLink(fields, &result) != nil || result.RequestID != input.RequestID || result.ExpiresAt <= 0 {
		return DeviceLinkApproval{}, linkFailure("BOOTSTRAP_INVALID_RESPONSE")
	}
	if input.ExpectedExpiresAt != "" {
		previous, err := strconv.ParseInt(input.ExpectedExpiresAt, 10, 64)
		if err != nil || result.State != "pending" || result.ExpiresAt <= previous {
			return DeviceLinkApproval{}, linkFailure("BOOTSTRAP_INVALID_RESPONSE")
		}
	}
	switch result.State {
	case "pending", "expired", "expired_pending", "issued", "prepared", "complete":
		return result, nil
	}
	return DeviceLinkApproval{}, linkFailure("BOOTSTRAP_INVALID_RESPONSE")
}
