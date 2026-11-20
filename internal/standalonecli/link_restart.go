package standalonecli

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
)

func (f *linkFlow) validateRestart(ctx context.Context) error {
	if !f.interactive {
		return linkError("LINK_APPROVAL_REQUIRED")
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
	return f.checkCurrent(ctx)
}

// RestartPendingLink is called only after explicit restart and terminal server
// evidence. The lock and exact custody comparison prevent a stale process from
// replacing a newer request or resurrecting a logged-out session.
func (s *ProtectedStore) RestartPendingLink(ctx context.Context, p PendingLink) (PendingLink, error) {
	if !p.valid() {
		return PendingLink{}, securecredential.ErrUnavailable
	}
	unlock, record, err := s.lockLinkSession(ctx, p.LinkBinding)
	if err != nil {
		return PendingLink{}, err
	}
	defer unlock()
	if record.PendingLink == nil || !record.PendingLink.sameRequest(p) {
		return PendingLink{}, securecredential.ErrUnavailable
	}
	id, secret := make([]byte, 32), make([]byte, 32)
	if _, err = rand.Read(id); err != nil {
		return PendingLink{}, securecredential.ErrUnavailable
	}
	if _, err = rand.Read(secret); err != nil {
		return PendingLink{}, securecredential.ErrUnavailable
	}
	next := PendingLink{LinkBinding: p.LinkBinding, Phase: LinkPrepared, RequestID: base64.RawURLEncoding.EncodeToString(id), RecoverySecret: base64.RawURLEncoding.EncodeToString(secret)}
	record.PendingLink = &next
	if err = s.saveRecord(ctx, record); err != nil {
		return PendingLink{}, err
	}
	return next, nil
}
