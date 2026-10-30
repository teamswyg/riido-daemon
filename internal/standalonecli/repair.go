package standalonecli

import (
	"context"
	"errors"
	"net/http"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type repairStore interface {
	BeginPendingRepair(context.Context, string) (PendingRepair, error)
	LoadPendingRepair(context.Context) (*PendingRepair, error)
	MarkPendingRepairCreated(context.Context, PendingRepair) (PendingRepair, error)
	RestartPendingRepair(context.Context, PendingRepair) (PendingRepair, error)
	ClearPendingRepair(context.Context, PendingRepair) error
	ExistingInstallationKey(context.Context, string) (userauth.InstallationKey, error)
	SignInstallationRepairProof(context.Context, userauth.InstallationKey, userauth.InstallationRepairProof) (string, error)
	SignDeviceLinkProof(context.Context, userauth.InstallationKey, userauth.DeviceLinkProof) (string, error)
}

// Repair never calls the OAuth manager or creates an installation key. A fresh
// browser decision authorizes only the original installation's device handoff.
func (s *Service) Repair(ctx context.Context, daemon LinkDaemon, interactive, restart bool) (LinkResult, error) {
	store, ok := s.store.(repairStore)
	if !ok {
		return LinkResult{}, securecredential.ErrUnavailable
	}
	if restart && !interactive {
		return LinkResult{}, repairLocalError("REPAIR_APPROVAL_REQUIRED")
	}
	p, err := store.LoadPendingRepair(ctx)
	if err != nil {
		return LinkResult{}, err
	}
	if p == nil {
		if !interactive {
			return LinkResult{}, repairLocalError("REPAIR_APPROVAL_REQUIRED")
		}
		key, err := store.ExistingInstallationKey(ctx, s.origin)
		if err != nil {
			return LinkResult{}, err
		}
		pending, err := store.BeginPendingRepair(ctx, key.KeyFingerprint)
		if err != nil {
			return LinkResult{}, err
		}
		p = &pending
		restart = false
	}
	f := repairFlow{service: s, store: store, daemon: daemon, pending: *p, interactive: interactive}
	if err = f.run(ctx, restart); err != nil {
		return LinkResult{}, err
	}
	return LinkResult{Mode: "user", APIOrigin: s.origin, Linked: true}, nil
}

func repairRejected(err error) bool {
	var remote *userauth.InstallationRepairError
	return errors.As(err, &remote) && remote.StatusCode == http.StatusBadRequest && remote.Code == "REPAIR_NOT_FOUND"
}

func repairCreateAllowed(err error, p PendingRepair) bool {
	return p.Phase == LinkPrepared && repairRejected(err)
}
