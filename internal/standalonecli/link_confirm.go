package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func (f *linkFlow) confirm(ctx context.Context, proof *userauth.DeviceLinkProof) error {
	if !f.interactive {
		return linkError("LINK_APPROVAL_REQUIRED")
	}
	if proof == nil {
		return linkError("LINK_INVALID_RESPONSE")
	}
	if f.daemon == nil {
		return linkError("LINK_DAEMON_UNAVAILABLE")
	}
	device, err := f.daemon.Info(ctx)
	if err != nil {
		return err
	}
	if device.Resource != f.pending.Resource || device.DeviceID != f.pending.ExpectedDeviceID {
		return linkError("LINK_BINDING_MISMATCH")
	}
	key, err := f.store.ExistingInstallationKey(ctx, f.pending.Resource)
	if err != nil {
		return err
	}
	if key.KeyFingerprint != f.pending.KeyFingerprint {
		return linkError("LINK_BINDING_MISMATCH")
	}
	if err := f.checkCurrent(ctx); err != nil {
		return err
	}
	signature, err := f.store.SignDeviceLinkProof(ctx, key, *proof)
	if err != nil {
		return err
	}
	if err := f.checkCurrent(ctx); err != nil {
		return err
	}
	err = f.daemon.Confirm(ctx, *proof, signature)
	f.confirmed = true
	if err == nil {
		return nil
	}
	// A lost daemon ACK may follow a committed server receipt. Recover the
	// original pair once before reporting the transport error; never recreate.
	result, recoverErr := f.recover(ctx)
	if recoverErr == nil && result.State == "linked" {
		// The main loop performs session-bound cleanup after another receipt
		// read, keeping one success path even for an ambiguous ACK.
		return nil
	}
	return err
}
