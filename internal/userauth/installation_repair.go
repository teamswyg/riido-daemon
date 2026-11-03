package userauth

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

type InstallationRepairProof struct {
	Resource       string `json:"resource"`
	ClientID       string `json:"client_id"`
	RequestID      string `json:"request_id"`
	KeyFingerprint string `json:"key_fingerprint"`
	RecoveryHash   string `json:"recovery_hash"`
}
type InstallationRepairRequest struct {
	RequestID, RecoveryHash, Signature string
	Key                                InstallationKey
}
type (
	InstallationRepairRecoveryRequest struct{ RequestID, RecoverySecret, KeyFingerprint string }
	InstallationRepairApproval        struct {
		RequestID     string `json:"request_id"`
		State         string `json:"state"`
		ExpiresAt     int64  `json:"expires_at"`
		RecoveryUntil int64  `json:"recovery_until"`
	}
)

type InstallationRepairError struct {
	Code       string
	StatusCode int
	Retryable  bool
}

func (*InstallationRepairError) Error() string       { return "CLI installation repair failed" }
func (InstallationRepairProof) String() string       { return "[redacted installation repair proof]" }
func (p InstallationRepairProof) GoString() string   { return p.String() }
func (InstallationRepairRequest) String() string     { return "[redacted installation repair request]" }
func (p InstallationRepairRequest) GoString() string { return p.String() }
func (InstallationRepairRecoveryRequest) String() string {
	return "[redacted installation repair recovery]"
}
func (p InstallationRepairRecoveryRequest) GoString() string { return p.String() }
func (InstallationRepairApproval) String() string            { return "[redacted installation repair approval]" }
func (p InstallationRepairApproval) GoString() string        { return p.String() }

// CreateInstallationRepair carries only an existing-key proof. It cannot read
// or refresh an OAuth grant. The caller retains the request pair before I/O.
func (c *Client) CreateInstallationRepair(ctx context.Context, input InstallationRepairRequest) (InstallationRepairApproval, error) {
	if !validDeviceLinkClient(c) || !deviceLinkRandom(input.RequestID) || !deviceLinkDigest(input.RecoveryHash) || !validInstallationSignature(input.Signature) || c.validateInstallationConfiguration(input.Key, func(context.Context, InstallationProof) (string, error) { return "", nil }) != nil {
		return InstallationRepairApproval{}, repairFailure("REPAIR_INVALID_CONFIG")
	}
	raw, status, err := c.deviceLinkPOST(ctx, "/oauth/cli/installation-repairs", map[string]string{"request_id": input.RequestID, "resource": c.config.Resource, "client_id": c.config.ClientID, "key_algorithm": input.Key.Algorithm, "public_key": input.Key.PublicKey, "recovery_hash": input.RecoveryHash, "signature": input.Signature}, "")
	if err != nil {
		return InstallationRepairApproval{}, repairTransportError(err)
	}
	if status != 201 {
		return InstallationRepairApproval{}, repairHTTPError(status, raw)
	}
	fields, err := deviceLinkObject(raw, "request_id", "state", "expires_at", "recovery_until")
	var result InstallationRepairApproval
	if err != nil || decodeDeviceLink(fields, &result) != nil || result.RequestID != input.RequestID || (result.State != "pending" && result.State != "approved" && result.State != "rejected") || result.ExpiresAt <= 0 || result.RecoveryUntil < result.ExpiresAt {
		return InstallationRepairApproval{}, repairFailure("REPAIR_INVALID_RESPONSE")
	}
	return result, nil
}

// RecoverInstallationRepair uses request custody only. An approved original
// device comes from the bound server proof; the caller must compare it with the
// authenticated daemon before signing. A final receipt needs no live daemon.
func (c *Client) RecoverInstallationRepair(ctx context.Context, input InstallationRepairRecoveryRequest) (DeviceLinkRecoveryResult, error) {
	if !validDeviceLinkClient(c) || !deviceLinkRandom(input.RequestID) || !deviceLinkRandom(input.RecoverySecret) || !deviceLinkDigest(input.KeyFingerprint) {
		return DeviceLinkRecoveryResult{}, repairFailure("REPAIR_INVALID_CONFIG")
	}
	raw, status, err := c.deviceLinkPOST(ctx, "/oauth/cli/installation-repairs/recover", map[string]string{"request_id": input.RequestID, "recovery_secret": input.RecoverySecret}, "")
	if err != nil {
		return DeviceLinkRecoveryResult{}, repairTransportError(err)
	}
	if status != 200 {
		return DeviceLinkRecoveryResult{}, repairHTTPError(status, raw)
	}
	return c.decodeInstallationRepairRecovery(raw, input)
}

func (c *Client) decodeInstallationRepairRecovery(raw []byte, input InstallationRepairRecoveryRequest) (DeviceLinkRecoveryResult, error) {
	var shape DeviceLinkRecoveryResult
	if json.Unmarshal(raw, &shape) != nil {
		return shape, repairFailure("REPAIR_INVALID_RESPONSE")
	}
	if shape.State == "proof_required" {
		if shape.Proof == nil || !deviceLinkID(shape.Proof.ExpectedDeviceID) {
			return DeviceLinkRecoveryResult{}, repairFailure("REPAIR_INVALID_RESPONSE")
		}
		// Reuse the exact schema, origin/client/key and expiry checks. Duplicate
		// keys are checked against original raw bytes, never this preliminary shape.
		result, err := c.decodeDeviceLinkRecovery(raw, DeviceLinkRecoveryRequest{RequestID: input.RequestID, KeyFingerprint: input.KeyFingerprint, ExpectedDeviceID: shape.Proof.ExpectedDeviceID})
		return result, repairTransportError(err)
	}
	fields, err := deviceLinkObject(raw, "request_id", "state")
	var result DeviceLinkRecoveryResult
	if err != nil || decodeDeviceLink(fields, &result) != nil || result.RequestID != input.RequestID {
		return result, repairFailure("REPAIR_INVALID_RESPONSE")
	}
	switch result.State {
	case "awaiting_approval", "preparing", "approval_expired", "repair_expired", "consent_rejected", "recovery_expired", "linked":
		return result, nil
	}
	return DeviceLinkRecoveryResult{}, repairFailure("REPAIR_INVALID_RESPONSE")
}

func repairFailure(code string) error { return &InstallationRepairError{Code: code} }
func repairTransportError(err error) error {
	var link *DeviceLinkError
	if errors.As(err, &link) {
		return &InstallationRepairError{Code: strings.Replace(link.Code, "LINK_", "REPAIR_", 1), StatusCode: link.StatusCode, Retryable: link.Retryable}
	}
	return err
}

func repairHTTPError(status int, raw []byte) error {
	code := "REPAIR_REQUEST_REJECTED"
	retry := status == 408 || status == 429 || status >= 500
	if retry {
		code = "REPAIR_HTTP_ERROR"
	}
	if status == 429 {
		code = "REPAIR_RATE_LIMITED"
	}
	if fields, err := deviceLinkObject(raw, "error", "error_description"); err == nil {
		var serverCode string
		if json.Unmarshal(fields["error"], &serverCode) == nil {
			if status == 400 && serverCode == "invalid_installation_repair" {
				code = "REPAIR_NOT_FOUND"
			}
			if serverCode == "installation_policy_unknown" {
				code = "REPAIR_POLICY_UNKNOWN"
			}
		}
	}
	return &InstallationRepairError{Code: code, StatusCode: status, Retryable: retry}
}
