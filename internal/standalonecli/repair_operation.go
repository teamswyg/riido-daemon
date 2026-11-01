package standalonecli

import (
	"context"
	"errors"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
)

// LockRepairOperation uses the separate origin-bound repair item lock. The
// credential lock remains short-lived so logout can invalidate pending custody.
func (s *ProtectedStore) LockRepairOperation(ctx context.Context) (func(), error) {
	if s == nil || s.RepairNative == nil || !validLinkResource(s.Resource) {
		return nil, securecredential.ErrUnavailable
	}
	unlock, err := s.RepairNative.TryLock(ctx)
	if errors.Is(err, securecredential.ErrBusy) {
		return nil, repairLocalError("REPAIR_IN_PROGRESS")
	}
	if err == nil && unlock == nil {
		return nil, securecredential.ErrUnavailable
	}
	return unlock, mapStorageError(err)
}
