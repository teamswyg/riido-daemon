package standalonecli

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
)

// RestartPendingRepair is only called by explicit interactive restart after a
// terminal server result. It atomically replaces exact custody, never a key.
func (s *ProtectedStore) RestartPendingRepair(ctx context.Context, p PendingRepair) (PendingRepair, error) {
	unlock, current, err := s.lockExactRepair(ctx, p)
	if err != nil {
		return PendingRepair{}, err
	}
	defer unlock()
	id, secret := make([]byte, 32), make([]byte, 32)
	if _, err = rand.Read(id); err != nil {
		return PendingRepair{}, securecredential.ErrUnavailable
	}
	if _, err = rand.Read(secret); err != nil {
		return PendingRepair{}, securecredential.ErrUnavailable
	}
	current.RequestID = base64.RawURLEncoding.EncodeToString(id)
	current.RecoverySecret = base64.RawURLEncoding.EncodeToString(secret)
	current.Phase = LinkPrepared
	if err = s.saveRepair(ctx, current); err != nil {
		return PendingRepair{}, err
	}
	return current, nil
}
