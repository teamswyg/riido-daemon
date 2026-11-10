package userauth

import "context"

func (c *Client) CreateTransferRepair(ctx context.Context, original string, r InstallationRepairRequest) error {
	if !validDeviceLinkClient(c) || !deviceLinkRandom(original) || !deviceLinkRandom(r.RequestID) || original == r.RequestID || !deviceLinkDigest(r.RecoveryHash) || !validInstallationSignature(r.Signature) || c.validateInstallationConfiguration(r.Key, func(context.Context, InstallationProof) (string, error) { return "", nil }) != nil {
		return repairFailure("TRANSFER_INVALID_CONFIG")
	}
	raw, status, err := c.deviceLinkPOST(ctx, transferPath+"/repairs", map[string]string{"original_request_id": original, "request_id": r.RequestID, "resource": c.config.Resource, "client_id": c.config.ClientID, "key_algorithm": r.Key.Algorithm, "public_key": r.Key.PublicKey, "recovery_hash": r.RecoveryHash, "signature": r.Signature}, "")
	if err != nil {
		return err
	}
	if status != 201 {
		return repairHTTPError(status, raw)
	}
	fields, err := deviceLinkObject(raw, "request_id", "state", "expires_at", "recovery_until")
	var a InstallationRepairApproval
	if err != nil || decodeDeviceLink(fields, &a) != nil || a.RequestID != r.RequestID || a.ExpiresAt <= 0 || a.RecoveryUntil < a.ExpiresAt {
		return repairFailure("TRANSFER_INVALID_RESPONSE")
	}
	switch a.State {
	case "pending", "approved", "rejected", "expired":
		return nil
	}
	return repairFailure("TRANSFER_INVALID_RESPONSE")
}

func (c *Client) RecoverTransferRepair(ctx context.Context, id, secret string) (string, error) {
	if !validDeviceLinkClient(c) || !deviceLinkRandom(id) || !deviceLinkRandom(secret) {
		return "", repairFailure("TRANSFER_INVALID_CONFIG")
	}
	raw, status, err := c.deviceLinkPOST(ctx, transferPath+"/repairs/recover", map[string]string{"request_id": id, "recovery_secret": secret}, "")
	if err != nil {
		return "", err
	}
	if status != 200 {
		return "", repairHTTPError(status, raw)
	}
	fields, err := deviceLinkObject(raw, "request_id", "state")
	var result TransferResult
	if err != nil || decodeDeviceLink(fields, &result) != nil || result.RequestID != id {
		return "", repairFailure("TRANSFER_INVALID_RESPONSE")
	}
	switch result.State {
	case "pending", "expired", "rejected", "awaiting_evidence", "awaiting_receipt", "complete", "retired":
		return result.State, nil
	}
	return "", repairFailure("TRANSFER_INVALID_RESPONSE")
}
