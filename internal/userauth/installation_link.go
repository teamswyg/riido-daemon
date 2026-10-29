package userauth

import (
	"context"
	"strings"
)

// CreateDeviceLink is a fixed mutation, not an extension of generic Do. The
// expected session belongs to the already protected pending request.
func (m *Manager) CreateDeviceLink(ctx context.Context, expectedSessionID string, input DeviceLinkRequest) (DeviceLinkApproval, error) {
	if m == nil || !validDeviceLinkClient(m.client) || !deviceLinkID(expectedSessionID) || !validDeviceLinkRequest(input) {
		return DeviceLinkApproval{}, linkFailure("LINK_INVALID_CONFIG")
	}
	generation := m.generation.Load()
	credential, err := m.accessCredential(ctx, "", expectedSessionID)
	if err != nil {
		return DeviceLinkApproval{}, err
	}
	if m.generation.Load() != generation {
		return DeviceLinkApproval{}, ErrLoggedOut
	}
	if credential.TokenType != "Bearer" || len(strings.Fields(credential.Scope)) != 1 || strings.Fields(credential.Scope)[0] != "read" {
		return DeviceLinkApproval{}, linkFailure("LINK_INVALID_CONFIG")
	}
	raw, status, err := m.client.deviceLinkPOST(ctx, "/oauth/cli/installation-links", map[string]string{"request_id": input.RequestID, "expected_device_id": input.ExpectedDeviceID, "recovery_secret": input.RecoverySecret}, credential.AccessToken)
	if err != nil {
		return DeviceLinkApproval{}, err
	}
	if err = m.checkDeviceLinkSession(ctx, expectedSessionID, generation); err != nil {
		return DeviceLinkApproval{}, err
	}
	if status != 201 {
		return DeviceLinkApproval{}, deviceLinkHTTPError(status, raw, false)
	}
	fields, err := deviceLinkObject(raw, "request_id", "state", "expires_at")
	if err != nil {
		return DeviceLinkApproval{}, err
	}
	var result DeviceLinkApproval
	if decodeDeviceLink(fields, &result) != nil || result.RequestID != input.RequestID || result.ExpiresAt <= 0 || (result.State != "pending" && result.State != "approved") {
		return DeviceLinkApproval{}, linkFailure("LINK_INVALID_RESPONSE")
	}
	// An exact server retry may preserve an already expired approval window.
	// This response never authorizes a proof or renews its original deadline.
	return result, nil
}

func (m *Manager) checkDeviceLinkSession(ctx context.Context, expected string, generation uint64) error {
	if m.generation.Load() != generation {
		return ErrLoggedOut
	}
	unlock, err := m.lock(ctx)
	if err != nil {
		return err
	}
	defer unlock()
	current, err := m.store.Load(ctx)
	if err != nil {
		return storageError(err)
	}
	if m.generation.Load() != generation || current.SessionID != expected {
		return ErrLoggedOut
	}
	return nil
}

// RecoverDeviceLink uses request-scoped custody only. It does not load, refresh,
// replace or revoke an OAuth credential; the server checks the original grant.
func (c *Client) RecoverDeviceLink(ctx context.Context, input DeviceLinkRecoveryRequest) (DeviceLinkRecoveryResult, error) {
	if !validDeviceLinkClient(c) || !validDeviceLinkRequest(DeviceLinkRequest{RequestID: input.RequestID, RecoverySecret: input.RecoverySecret, ExpectedDeviceID: input.ExpectedDeviceID}) || !deviceLinkDigest(input.KeyFingerprint) {
		return DeviceLinkRecoveryResult{}, linkFailure("LINK_INVALID_CONFIG")
	}
	raw, status, err := c.deviceLinkPOST(ctx, "/oauth/cli/installation-links/recover", map[string]string{"request_id": input.RequestID, "recovery_secret": input.RecoverySecret}, "")
	if err != nil {
		return DeviceLinkRecoveryResult{}, err
	}
	if status != 200 {
		return DeviceLinkRecoveryResult{}, deviceLinkHTTPError(status, raw, true)
	}
	return c.decodeDeviceLinkRecovery(raw, input)
}
