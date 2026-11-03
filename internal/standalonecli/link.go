package standalonecli

import (
	"context"
	"errors"
	"net/http"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type LinkDevice struct{ Resource, DeviceID string }

// LinkDaemon accepts neither OAuth credentials nor recovery custody.
type LinkDaemon interface {
	Info(context.Context) (LinkDevice, error)
	Confirm(context.Context, userauth.DeviceLinkProof, string) error
}

type linkStore interface {
	BeginPendingLink(context.Context, LinkBinding) (PendingLink, error)
	LoadPendingLink(context.Context, string) (*PendingLink, error)
	MarkPendingLinkCreated(context.Context, PendingLink) (PendingLink, error)
	ClearPendingLink(context.Context, PendingLink) error
	ExistingInstallationKey(context.Context, string) (userauth.InstallationKey, error)
	SignDeviceLinkProof(context.Context, userauth.InstallationKey, userauth.DeviceLinkProof) (string, error)
}

type LinkResult struct {
	Mode      string `json:"mode"`
	APIOrigin string `json:"api_origin"`
	Linked    bool   `json:"linked"`
}

func linkError(code string) error { return &userauth.DeviceLinkError{Code: code} }

// Link is explicit: ordinary login/read/refresh never calls the daemon.
// Errors leave the original request protected for a later invocation.
func (s *Service) Link(ctx context.Context, daemon LinkDaemon, interactive bool) (LinkResult, error) {
	store, ok := s.store.(linkStore)
	if !ok {
		return LinkResult{}, securecredential.ErrUnavailable
	}
	sessionID, err := s.linkSession(ctx)
	if err != nil {
		return LinkResult{}, err
	}
	pending, err := store.LoadPendingLink(ctx, sessionID)
	if err != nil {
		return LinkResult{}, err
	}
	if pending == nil {
		if !interactive {
			return LinkResult{}, linkError("LINK_APPROVAL_REQUIRED")
		}
		pending, err = s.prepareLink(ctx, store, daemon, sessionID)
		if err != nil {
			return LinkResult{}, err
		}
	}
	flow := linkFlow{service: s, store: store, daemon: daemon, pending: *pending, interactive: interactive}
	if err := flow.run(ctx); err != nil {
		return LinkResult{}, err
	}
	return LinkResult{Mode: "user", APIOrigin: s.origin, Linked: true}, nil
}

func (s *Service) linkSession(ctx context.Context) (string, error) {
	unlock, err := s.store.Lock(ctx)
	if err != nil {
		return "", err
	}
	defer unlock()
	credential, err := s.store.Load(ctx)
	if err != nil {
		return "", err
	}
	if !linkBindingID(credential.SessionID) {
		return "", userauth.ErrNotAuthenticated
	}
	return credential.SessionID, nil
}

func (s *Service) prepareLink(ctx context.Context, store linkStore, daemon LinkDaemon, sessionID string) (*PendingLink, error) {
	if daemon == nil {
		return nil, linkError("LINK_DAEMON_UNAVAILABLE")
	}
	device, err := daemon.Info(ctx)
	if err != nil {
		return nil, err
	}
	if device.Resource != s.origin || !linkBindingID(device.DeviceID) {
		return nil, linkError("LINK_BINDING_MISMATCH")
	}
	key, err := store.ExistingInstallationKey(ctx, s.origin)
	if err != nil {
		return nil, err
	}
	pending, err := store.BeginPendingLink(ctx, LinkBinding{Resource: s.origin, ClientID: "riido-cli", SessionID: sessionID, KeyFingerprint: key.KeyFingerprint, ExpectedDeviceID: device.DeviceID})
	return &pending, err
}

func linkCreateAllowed(err error, pending PendingLink) bool {
	var failure *userauth.DeviceLinkError
	return pending.Phase == LinkPrepared && errors.As(err, &failure) && failure.StatusCode == http.StatusBadRequest && failure.Code == "LINK_RECOVERY_REJECTED"
}
