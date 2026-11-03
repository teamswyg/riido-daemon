package standalonecli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type repairFlow struct {
	service             *Service
	store               repairStore
	daemon              LinkDaemon
	pending             PendingRepair
	interactive, opened bool
	confirmedIntent     string
}

func (f *repairFlow) recover(ctx context.Context) (userauth.DeviceLinkRecoveryResult, error) {
	p := f.pending
	return f.service.client.RecoverInstallationRepair(ctx, userauth.InstallationRepairRecoveryRequest{RequestID: p.RequestID, RecoverySecret: p.RecoverySecret, KeyFingerprint: p.KeyFingerprint})
}

func (f *repairFlow) checkCurrent(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	p, err := f.store.LoadPendingRepair(ctx)
	if err != nil {
		return err
	}
	if p == nil || !p.sameRequest(f.pending) {
		return repairLocalError("REPAIR_REQUEST_CHANGED")
	}
	f.pending = *p
	return nil
}

func (f *repairFlow) key(ctx context.Context) (userauth.InstallationKey, error) {
	key, err := f.store.ExistingInstallationKey(ctx, f.pending.Resource)
	if err != nil {
		return key, err
	}
	if key.KeyFingerprint != f.pending.KeyFingerprint {
		return userauth.InstallationKey{}, repairLocalError("REPAIR_BINDING_MISMATCH")
	}
	return key, f.checkCurrent(ctx)
}

func (f *repairFlow) run(ctx context.Context, restart bool) error {
	if err := f.checkCurrent(ctx); err != nil {
		return err
	}
	result, err := f.recover(ctx)
	if restart && result.State != "linked" {
		// A generic 400 can mean a temporarily inactive client or principal.
		// Only positive terminal evidence permits discarding the original pair.
		terminal := err == nil && (result.State == "approval_expired" || result.State == "repair_expired" || result.State == "consent_rejected" || result.State == "recovery_expired")
		if !terminal {
			if err != nil {
				return err
			}
			return repairLocalError("REPAIR_IN_PROGRESS")
		}
		if _, err = f.key(ctx); err != nil {
			return err
		}
		f.pending, err = f.store.RestartPendingRepair(ctx, f.pending)
		if err != nil {
			return err
		}
		result, err = f.recover(ctx)
	}
	if repairCreateAllowed(err, f.pending) {
		if currentErr := f.checkCurrent(ctx); currentErr != nil {
			return currentErr
		}
		if !repairCreateAllowed(err, f.pending) {
			return err
		}
		if err = f.create(ctx); err != nil {
			return err
		}
		result, err = f.recover(ctx)
	}
	for err == nil {
		if result.State == "linked" {
			return f.store.ClearPendingRepair(ctx, f.pending)
		}
		if f.pending.Phase != LinkCreated {
			f.pending, err = f.store.MarkPendingRepairCreated(ctx, f.pending)
			if err != nil {
				return err
			}
		}
		if err = f.advance(ctx, result); err != nil {
			return err
		}
		if err = waitLinkPoll(ctx); err != nil {
			return err
		}
		result, err = f.recover(ctx)
	}
	return err
}

func (f *repairFlow) create(ctx context.Context) error {
	if !f.interactive {
		return repairLocalError("REPAIR_APPROVAL_REQUIRED")
	}
	key, err := f.key(ctx)
	if err != nil {
		return err
	}
	p := f.pending
	hash := sha256.Sum256([]byte(p.RecoverySecret))
	proof := userauth.InstallationRepairProof{Resource: p.Resource, ClientID: p.ClientID, RequestID: p.RequestID, KeyFingerprint: p.KeyFingerprint, RecoveryHash: hex.EncodeToString(hash[:])}
	signature, err := f.store.SignInstallationRepairProof(ctx, key, proof)
	if err != nil {
		return err
	}
	if err = f.checkCurrent(ctx); err != nil {
		return err
	}
	_, err = f.service.client.CreateInstallationRepair(ctx, userauth.InstallationRepairRequest{RequestID: p.RequestID, RecoveryHash: proof.RecoveryHash, Signature: signature, Key: key})
	if err != nil {
		return err
	}
	f.pending, err = f.store.MarkPendingRepairCreated(ctx, f.pending)
	return err
}

func (f *repairFlow) advance(ctx context.Context, result userauth.DeviceLinkRecoveryResult) error {
	switch result.State {
	case "preparing":
		return nil
	case "approval_expired":
		return repairLocalError("REPAIR_APPROVAL_EXPIRED")
	case "repair_expired":
		return repairLocalError("REPAIR_EXPIRED")
	case "consent_rejected":
		return repairLocalError("REPAIR_CONSENT_REJECTED")
	case "recovery_expired":
		return repairLocalError("REPAIR_RECOVERY_EXPIRED")
	case "awaiting_approval":
		if f.opened {
			return nil
		}
		if !f.interactive {
			return repairLocalError("REPAIR_APPROVAL_REQUIRED")
		}
		if f.service.openBrowser == nil {
			return ErrBrowserUnavailable
		}
		if err := f.checkCurrent(ctx); err != nil {
			return err
		}
		address := f.pending.Resource + "/oauth/cli/installation-repairs/" + f.pending.RequestID + "/authorize"
		if err := f.service.openBrowser(ctx, address); err != nil {
			return ErrBrowserUnavailable
		}
		f.opened = true
		return nil
	case "proof_required":
		if result.Proof == nil {
			return repairLocalError("REPAIR_INVALID_RESPONSE")
		}
		if f.confirmedIntent == result.Proof.IntentID {
			return nil
		}
		return f.confirm(ctx, result.Proof)
	}
	return repairLocalError("REPAIR_INVALID_RESPONSE")
}

func (f *repairFlow) confirm(ctx context.Context, proof *userauth.DeviceLinkProof) error {
	if !f.interactive {
		return repairLocalError("REPAIR_APPROVAL_REQUIRED")
	}
	if f.daemon == nil {
		return repairLocalError("REPAIR_DAEMON_UNAVAILABLE")
	}
	device, err := f.daemon.Info(ctx)
	if err != nil {
		return err
	}
	if device.Resource != f.pending.Resource || device.DeviceID != proof.ExpectedDeviceID {
		return repairLocalError("REPAIR_BINDING_MISMATCH")
	}
	key, err := f.key(ctx)
	if err != nil {
		return err
	}
	signature, err := f.store.SignDeviceLinkProof(ctx, key, *proof)
	if err != nil {
		return err
	}
	if err = f.checkCurrent(ctx); err != nil {
		return err
	}
	err = f.daemon.Confirm(ctx, *proof, signature)
	f.confirmedIntent = proof.IntentID
	if err == nil {
		return nil
	}
	result, recoverErr := f.recover(ctx)
	if recoverErr == nil && result.State == "linked" {
		return nil
	}
	return err
}
