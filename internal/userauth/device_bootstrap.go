package userauth

import (
	"context"
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
	DeviceID       string `json:"expected_device_id"`
	RequestID      string `json:"request_id"`
	RecoverySecret string `json:"recovery_secret"`
	NextHash       string `json:"next_secret_hash"`
}

func (BootstrapRequest) String() string     { return "[redacted device bootstrap request]" }
func (r BootstrapRequest) GoString() string { return r.String() }

func ParseBootstrapRequest(raw []byte) (BootstrapRequest, error) {
	var input BootstrapRequest
	fields, err := deviceLinkObject(raw, "resource", "client_id", "installation_id", "key_fingerprint", "owner_principal_id", "request_id", "expected_device_id", "recovery_secret", "next_secret_hash")
	if err != nil || decodeDeviceLink(fields, &input) != nil || !deviceLinkRandom(input.RequestID) || !deviceLinkRandom(input.RecoverySecret) || !deviceLinkDigest(input.NextHash) || !deviceLinkDigest(input.KeyFingerprint) || !deviceLinkID(input.InstallationID) || !deviceLinkID(input.OwnerID) || input.ClientID != "riido-cli" || len(input.DeviceID) != 36 || !strings.HasPrefix(input.DeviceID, "dev_") || strings.Trim(input.DeviceID[4:], "0123456789abcdef") != "" {
		return BootstrapRequest{}, linkFailure("BOOTSTRAP_INVALID_REQUEST")
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
	raw, status, err := m.installationSessionPost(ctx, session, deviceBootstrapPath, map[string]string{"request_id": input.RequestID, "expected_device_id": input.DeviceID, "recovery_secret": input.RecoverySecret, "next_secret_hash": input.NextHash, "installation_id": input.InstallationID, "key_fingerprint": input.KeyFingerprint})
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
	switch result.State {
	case "pending", "issued", "prepared", "complete":
		return result, nil
	}
	return DeviceLinkApproval{}, linkFailure("BOOTSTRAP_INVALID_RESPONSE")
}
