package standalonecli

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
)

// MarkPendingLinkCreated records server acknowledgement under the original
// session lock. An ambiguous save or repeated acknowledgement never replaces
// the original pair or downgrades an already-created request.
func (s *ProtectedStore) MarkPendingLinkCreated(ctx context.Context, pending PendingLink) (PendingLink, error) {
	if !pending.valid() {
		return PendingLink{}, securecredential.ErrUnavailable
	}
	unlock, record, err := s.lockLinkSession(ctx, pending.LinkBinding)
	if err != nil {
		return PendingLink{}, err
	}
	defer unlock()
	current := record.PendingLink
	if current == nil || !current.sameRequest(pending) {
		return PendingLink{}, securecredential.ErrUnavailable
	}
	if current.Phase == LinkCreated {
		return *current, nil
	}
	current.Phase = LinkCreated
	if err = s.saveRecord(ctx, record); err != nil {
		return PendingLink{}, err
	}
	return *current, nil
}
