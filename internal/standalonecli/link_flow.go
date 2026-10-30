package standalonecli

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type linkFlow struct {
	service     *Service
	store       linkStore
	daemon      LinkDaemon
	pending     PendingLink
	interactive bool
	opened      bool
	confirmed   bool
}

func (f *linkFlow) recover(ctx context.Context) (userauth.DeviceLinkRecoveryResult, error) {
	p := f.pending
	return f.service.client.RecoverDeviceLink(ctx, userauth.DeviceLinkRecoveryRequest{RequestID: p.RequestID, RecoverySecret: p.RecoverySecret, ExpectedDeviceID: p.ExpectedDeviceID, KeyFingerprint: p.KeyFingerprint})
}

func (f *linkFlow) checkCurrent(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	pending, err := f.store.LoadPendingLink(ctx, f.pending.SessionID)
	if err != nil {
		return err
	}
	if pending == nil || !pending.sameRequest(f.pending) {
		return linkError("LINK_REQUEST_CHANGED")
	}
	f.pending = *pending
	return nil
}

func (f *linkFlow) run(ctx context.Context) error {
	if err := f.checkCurrent(ctx); err != nil {
		return err
	}
	result, err := f.recover(ctx)
	if linkCreateAllowed(err, f.pending) {
		if currentErr := f.checkCurrent(ctx); currentErr != nil {
			return currentErr
		}
		if !linkCreateAllowed(err, f.pending) {
			return err
		}
		if !f.interactive {
			return linkError("LINK_APPROVAL_REQUIRED")
		}
		_, err = f.service.manager.CreateDeviceLink(ctx, f.pending.SessionID, userauth.DeviceLinkRequest{RequestID: f.pending.RequestID, RecoverySecret: f.pending.RecoverySecret, ExpectedDeviceID: f.pending.ExpectedDeviceID})
		if err == nil {
			f.pending, err = f.store.MarkPendingLinkCreated(ctx, f.pending)
		}
		if err == nil {
			result, err = f.recover(ctx)
		}
	}
	for err == nil {
		if result.State == "linked" {
			return f.store.ClearPendingLink(ctx, f.pending)
		}
		if f.pending.Phase != LinkCreated {
			f.pending, err = f.store.MarkPendingLinkCreated(ctx, f.pending)
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

func (f *linkFlow) advance(ctx context.Context, result userauth.DeviceLinkRecoveryResult) error {
	switch result.State {
	case "proof_expired":
		return linkError("LINK_PROOF_EXPIRED")
	case "awaiting_approval":
		if f.opened {
			return nil
		}
		if !f.interactive {
			return linkError("LINK_APPROVAL_REQUIRED")
		}
		if f.service.openBrowser == nil {
			return ErrBrowserUnavailable
		}
		if err := f.checkCurrent(ctx); err != nil {
			return err
		}
		// Request IDs are validated canonical base64url; no server URL or secret
		// is accepted here. Only a human browser session can approve the request.
		address := f.pending.Resource + "/oauth/cli/installation-links/" + f.pending.RequestID + "/authorize"
		if err := f.service.openBrowser(ctx, address); err != nil {
			return ErrBrowserUnavailable
		}
		f.opened = true
		return nil
	case "proof_required":
		if f.confirmed {
			return nil
		}
		return f.confirm(ctx, result.Proof)
	default:
		return linkError("LINK_INVALID_RESPONSE")
	}
}

func waitLinkPoll(ctx context.Context) error {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
