package standalonecli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type TransferDaemon interface {
	Info(context.Context) (LinkDevice, error)
	ConfirmTransfer(context.Context, userauth.TransferProof, string) error
}
type transferSigner interface {
	SignDeviceTransferProof(context.Context, securecredential.InstallationPublicKey, securecredential.DeviceTransferProof) (string, error)
	SignInstallationTransferRepairProof(context.Context, securecredential.InstallationPublicKey, securecredential.InstallationTransferRepairProof) (string, error)
}

// State is the server's observation. Only complete means the device binding was published.
type TransferStatus struct {
	State string `json:"state"`
}

type transferFlow struct {
	s           *Service
	store       *ProtectedStore
	daemon      TransferDaemon
	journal     *transferJournal
	interactive bool
	opened      bool
	confirmed   string
}

func (s *Service) Transfer(ctx context.Context, daemon TransferDaemon, interactive, repair, restart bool) (TransferStatus, error) {
	if restart && (!repair || !interactive) {
		return TransferStatus{}, linkError("TRANSFER_INTERACTION_REQUIRED")
	}
	store, ok := s.store.(*ProtectedStore)
	if !ok || store.TransferNative == nil {
		return TransferStatus{}, securecredential.ErrUnavailable
	}
	unlock, err := store.LockRepairOperation(ctx)
	if err != nil {
		return TransferStatus{}, err
	}
	defer unlock()
	j, err := store.loadTransfer(ctx)
	if err != nil {
		return TransferStatus{}, err
	}
	f := transferFlow{s: s, daemon: daemon, store: store, journal: j, interactive: interactive}
	if j == nil {
		if !interactive || repair || restart {
			return TransferStatus{}, linkError("TRANSFER_NO_REQUEST")
		}
		if err = f.begin(ctx); err != nil {
			return TransferStatus{}, err
		}
	}
	if repair {
		return f.repair(ctx, restart)
	}
	return f.execute(ctx)
}

func (f *transferFlow) begin(ctx context.Context) error {
	if f.daemon == nil {
		return linkError("LINK_DAEMON_UNAVAILABLE")
	}
	d, err := f.daemon.Info(ctx)
	if err != nil {
		return err
	}
	if d.Resource != f.s.origin || !linkBindingID(d.DeviceID) {
		return linkError("LINK_BINDING_MISMATCH")
	}
	sid, err := f.s.linkSession(ctx)
	if err != nil {
		return err
	}
	key, err := f.store.ExistingInstallationKey(ctx, f.s.origin)
	if err != nil {
		return err
	}
	id, secret, err := newTransferPair()
	if err != nil {
		return err
	}
	f.journal = &transferJournal{Schema: "riido-cli-transfer.v1", Resource: f.s.origin, Origin: f.s.authOrigin, Session: sid, Device: d.DeviceID, Fingerprint: key.KeyFingerprint, Request: id, Secret: secret}
	return f.store.saveTransfer(ctx, *f.journal)
}

func (f *transferFlow) open(ctx context.Context, path string) error {
	if f.opened {
		return nil
	}
	if f.s.openBrowser == nil {
		return ErrBrowserUnavailable
	}
	if err := f.s.openBrowser(ctx, f.s.origin+"/oauth/cli/installation-transfers/"+path+"/authorize"); err != nil {
		return ErrBrowserUnavailable
	}
	f.opened = true
	return nil
}

func (f *transferFlow) finish(ctx context.Context, state string) (TransferStatus, error) {
	if err := f.store.TransferNative.Delete(ctx); err != nil {
		return TransferStatus{}, mapStorageError(err)
	}
	return TransferStatus{State: state}, nil
}

func (f *transferFlow) execute(ctx context.Context) (TransferStatus, error) {
	j := f.journal
	req := userauth.DeviceLinkRequest{RequestID: j.Request, RecoverySecret: j.Secret, ExpectedDeviceID: j.Device}
	if !j.Created && f.interactive {
		_, err := f.s.manager.CreateTransfer(ctx, j.Session, req)
		if err != nil {
			return TransferStatus{}, err
		}
		j.Created = true
		if err = f.store.saveTransfer(ctx, *j); err != nil {
			return TransferStatus{}, err
		}
	}
	for {
		result, err := f.s.manager.ResumeTransfer(ctx, j.Session, req, j.Fingerprint, f.interactive && f.confirmed == "")
		if err != nil {
			return TransferStatus{}, err
		}
		switch result.State {
		case "complete", "retired":
			return f.finish(ctx, result.State)
		case "proof_stopped":
			return TransferStatus{State: result.State}, nil
		case "pending":
			if f.interactive {
				err = f.open(ctx, j.Request)
			}
		case "proof_required":
			err = f.confirm(ctx, result.Proof)
		}
		if err != nil {
			return TransferStatus{}, err
		}
		if !f.interactive {
			return TransferStatus{State: result.State}, nil
		}
		if err = waitLinkPoll(ctx); err != nil {
			return TransferStatus{}, err
		}
	}
}

func (f *transferFlow) confirm(ctx context.Context, p *userauth.TransferProof) error {
	if p == nil || f.daemon == nil {
		return linkError("LINK_INVALID_RESPONSE")
	}
	d, err := f.daemon.Info(ctx)
	if err != nil {
		return err
	}
	if d.Resource != f.journal.Resource || d.DeviceID != f.journal.Device {
		return linkError("LINK_BINDING_MISMATCH")
	}
	key, err := f.store.ExistingInstallationKey(ctx, f.s.origin)
	if err != nil {
		return err
	}
	signer, ok := f.store.Native.(transferSigner)
	if !ok || key.KeyFingerprint != f.journal.Fingerprint {
		return securecredential.ErrUnavailable
	}
	native := securecredential.DeviceTransferProof{DeviceLinkProof: securecredential.DeviceLinkProof(p.DeviceLinkProof), PreviousDeviceID: p.PreviousDeviceID, PreviousLinkRevision: p.PreviousLinkRevision}
	signature, err := signer.SignDeviceTransferProof(ctx, securecredential.InstallationPublicKey(key), native)
	if err != nil {
		return mapLinkKeyError(err)
	}
	sid, err := f.s.linkSession(ctx)
	if err != nil {
		return err
	}
	if sid != f.journal.Session {
		return userauth.ErrLoggedOut
	}
	if err = f.daemon.ConfirmTransfer(ctx, *p, signature); err != nil {
		return err
	}
	f.confirmed = p.IntentID
	return nil
}

func (f *transferFlow) newRepair(ctx context.Context) error {
	id, secret, err := newTransferPair()
	if err != nil {
		return err
	}
	j := f.journal
	j.RepairRequest, j.RepairSecret, j.RepairCreated = id, secret, false
	return f.store.saveTransfer(ctx, *j)
}

func (f *transferFlow) createRepair(ctx context.Context) error {
	j := f.journal
	key, err := f.store.ExistingInstallationKey(ctx, j.Resource)
	if err != nil {
		return err
	}
	signer, ok := f.store.Native.(transferSigner)
	if !ok || key.KeyFingerprint != j.Fingerprint {
		return securecredential.ErrUnavailable
	}
	digest := sha256.Sum256([]byte(j.RepairSecret))
	proof := securecredential.InstallationTransferRepairProof{Resource: j.Resource, ClientID: "riido-cli", OriginalRequestID: j.Request, RequestID: j.RepairRequest, KeyFingerprint: j.Fingerprint, RecoveryHash: hex.EncodeToString(digest[:])}
	signature, err := signer.SignInstallationTransferRepairProof(ctx, securecredential.InstallationPublicKey(key), proof)
	if err != nil {
		return mapLinkKeyError(err)
	}
	if err = f.s.client.CreateTransferRepair(ctx, j.Request, userauth.InstallationRepairRequest{RequestID: j.RepairRequest, RecoveryHash: proof.RecoveryHash, Signature: signature, Key: key}); err != nil {
		return err
	}
	j.RepairCreated = true
	return f.store.saveTransfer(ctx, *j)
}

func (f *transferFlow) repair(ctx context.Context, restart bool) (TransferStatus, error) {
	j := f.journal
	if !j.Created {
		return TransferStatus{}, linkError("TRANSFER_NO_JOURNAL")
	}
	if j.RepairRequest == "" {
		if !f.interactive || restart {
			return TransferStatus{}, linkError("TRANSFER_NO_REPAIR")
		}
		if err := f.newRepair(ctx); err != nil {
			return TransferStatus{}, err
		}
	}
	for {
		if !j.RepairCreated && f.interactive && !restart {
			if err := f.createRepair(ctx); err != nil {
				return TransferStatus{}, err
			}
		}
		state, err := f.s.client.RecoverTransferRepair(ctx, j.RepairRequest, j.RepairSecret)
		if err != nil {
			return TransferStatus{}, err
		}
		switch state {
		case "complete", "retired":
			return f.finish(ctx, state)
		case "rejected", "expired":
			if !restart {
				return TransferStatus{State: state}, nil
			}
			if !f.interactive {
				return TransferStatus{}, linkError("TRANSFER_INTERACTION_REQUIRED")
			}
			if err = f.newRepair(ctx); err != nil {
				return TransferStatus{}, err
			}
			restart = false
			continue
		}
		if restart {
			return TransferStatus{}, linkError("TRANSFER_ACTIVE")
		}
		if !f.interactive || state == "awaiting_evidence" || state == "awaiting_receipt" {
			return TransferStatus{State: state}, nil
		}
		if state == "pending" {
			if err = f.open(ctx, "repairs/"+j.RepairRequest); err != nil {
				return TransferStatus{}, err
			}
		}
		if err = waitLinkPoll(ctx); err != nil {
			return TransferStatus{}, err
		}
	}
}
