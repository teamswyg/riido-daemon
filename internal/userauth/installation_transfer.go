package userauth

import (
	"context"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"
)

const transferPath = "/oauth/cli/installation-transfers"

type TransferProof struct {
	DeviceLinkProof
	PreviousDeviceID     string `json:"previous_device_id"`
	PreviousLinkRevision string `json:"previous_link_revision"`
}
type TransferResult struct {
	RequestID string         `json:"request_id"`
	State     string         `json:"state"`
	Proof     *TransferProof `json:"proof,omitempty"`
}

func (TransferResult) String() string     { return "[redacted device transfer]" }
func (r TransferResult) GoString() string { return r.String() }
func (TransferProof) String() string      { return "[redacted transfer proof]" }
func (r TransferProof) GoString() string  { return r.String() }

// All three live-grant calls are bound to the original local session. The
// fixed endpoints reject stale approvals without changing the OAuth login.
func (m *Manager) transferPost(ctx context.Context, session, path string, input DeviceLinkRequest) ([]byte, int, error) {
	if m == nil || !validDeviceLinkClient(m.client) || !validDeviceLinkRequest(input) || !deviceLinkID(session) {
		return nil, 0, linkFailure("TRANSFER_INVALID_CONFIG")
	}
	body := map[string]string{"request_id": input.RequestID, "recovery_secret": input.RecoverySecret}
	if path == transferPath {
		body["expected_device_id"] = input.ExpectedDeviceID
	}
	return m.installationSessionPost(ctx, session, path, body)
}

func (m *Manager) installationSessionPost(ctx context.Context, session, path string, body map[string]string) ([]byte, int, error) {
	if m == nil || !validDeviceLinkClient(m.client) || !deviceLinkID(session) {
		return nil, 0, linkFailure("LINK_INVALID_CONFIG")
	}
	gen := m.generation.Load()
	c, err := m.accessCredential(ctx, "", session)
	if err != nil {
		return nil, 0, err
	}
	if m.generation.Load() != gen {
		return nil, 0, ErrLoggedOut
	}
	if c.TokenType != "Bearer" || strings.TrimSpace(c.Scope) != "read" {
		return nil, 0, linkFailure("TRANSFER_INVALID_CONFIG")
	}
	raw, status, err := m.client.deviceLinkPOST(ctx, path, body, c.AccessToken)
	if err != nil {
		return nil, 0, err
	}
	if err = m.checkDeviceLinkSession(ctx, session, gen); err != nil {
		return nil, 0, err
	}
	return raw, status, nil
}

func (m *Manager) CreateTransfer(ctx context.Context, session string, input DeviceLinkRequest) (DeviceLinkApproval, error) {
	raw, status, err := m.transferPost(ctx, session, transferPath, input)
	if err != nil {
		return DeviceLinkApproval{}, err
	}
	if status != 201 {
		return DeviceLinkApproval{}, deviceLinkHTTPError(status, raw, false)
	}
	fields, err := deviceLinkObject(raw, "request_id", "state", "expires_at")
	var a DeviceLinkApproval
	if err != nil || decodeDeviceLink(fields, &a) != nil || a.RequestID != input.RequestID || a.ExpiresAt <= 0 || (a.State != "pending" && a.State != "approved") {
		return DeviceLinkApproval{}, linkFailure("TRANSFER_INVALID_RESPONSE")
	}
	return a, nil
}

func (m *Manager) ResumeTransfer(ctx context.Context, session string, input DeviceLinkRequest, fingerprint string, execute bool) (TransferResult, error) {
	if !deviceLinkDigest(fingerprint) {
		return TransferResult{}, linkFailure("TRANSFER_INVALID_CONFIG")
	}
	path := transferPath + "/recover"
	if execute {
		path = transferPath + "/execute"
	}
	raw, status, err := m.transferPost(ctx, session, path, input)
	if err != nil {
		return TransferResult{}, err
	}
	if status != 200 {
		return TransferResult{}, deviceLinkHTTPError(status, raw, true)
	}
	return m.client.decodeTransfer(raw, input, fingerprint, execute)
}

func (c *Client) decodeTransfer(raw []byte, input DeviceLinkRequest, fingerprint string, execute bool) (TransferResult, error) {
	bad := linkFailure("TRANSFER_INVALID_RESPONSE")
	var result TransferResult
	if json.Unmarshal(raw, &result) != nil || result.RequestID != input.RequestID {
		return TransferResult{}, bad
	}
	keys := []string{"request_id", "state"}
	switch result.State {
	case "pending", "not_started", "awaiting_evidence", "awaiting_receipt", "proof_stopped", "complete", "retired":
	case "proof_required":
		if !execute {
			return TransferResult{}, bad
		}
		keys = append(keys, "proof")
	default:
		return TransferResult{}, bad
	}
	fields, err := deviceLinkObject(raw, keys...)
	if err != nil {
		return TransferResult{}, bad
	}
	if result.Proof == nil {
		return result, nil
	}
	p := result.Proof
	_, err = deviceLinkObject(fields["proof"], "resource", "client_id", "installation_id", "intent_id", "challenge_hash", "key_fingerprint", "expected_device_id", "approved_principal_id", "expires_at", "previous_device_id", "previous_link_revision")
	revision, revErr := strconv.ParseUint(p.PreviousLinkRevision, 10, 64)
	expiry, timeErr := time.Parse(time.RFC3339Nano, p.ExpiresAt)
	if err != nil || p.Resource != c.config.Resource || p.ClientID != c.config.ClientID || p.ExpectedDeviceID != input.ExpectedDeviceID || p.KeyFingerprint != fingerprint || !deviceLinkRandom(p.IntentID) || !deviceLinkDigest(p.ChallengeHash) || !deviceLinkID(p.InstallationID) || !deviceLinkID(p.ApprovedPrincipalID) || !deviceLinkID(p.PreviousDeviceID) || p.PreviousDeviceID == p.ExpectedDeviceID || revErr != nil || revision == 0 || revision == math.MaxUint64 || strconv.FormatUint(revision, 10) != p.PreviousLinkRevision || timeErr != nil || expiry.UTC().Format(time.RFC3339Nano) != p.ExpiresAt || !expiry.After(time.Now()) {
		return TransferResult{}, bad
	}
	return result, nil
}
