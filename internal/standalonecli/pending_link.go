package standalonecli

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

// LinkBinding identifies the current local login and the explicit device/key
// selected for linking. It does not grant server authority on its own.
type LinkBinding struct {
	Resource         string `json:"resource"`
	ClientID         string `json:"client_id"`
	SessionID        string `json:"session_id"`
	KeyFingerprint   string `json:"key_fingerprint"`
	ExpectedDeviceID string `json:"expected_device_id"`
}

func (LinkBinding) String() string   { return "[redacted device link binding]" }
func (LinkBinding) GoString() string { return "[redacted device link binding]" }

// LinkPhase records progress, not request identity or server authority.
type LinkPhase string

const (
	LinkPrepared LinkPhase = "prepared"
	LinkCreated  LinkPhase = "created"
)

// PendingLink is secret material. It only becomes usable after the complete
// credential and pending record have been saved in the OS protected store.
type PendingLink struct {
	LinkBinding
	Phase          LinkPhase `json:"phase,omitempty"`
	RequestID      string    `json:"request_id"`
	RecoverySecret string    `json:"recovery_secret"`
}

func (PendingLink) String() string   { return "[redacted pending device link]" }
func (PendingLink) GoString() string { return "[redacted pending device link]" }

func (p PendingLink) sameRequest(other PendingLink) bool {
	return p.LinkBinding == other.LinkBinding && p.RequestID == other.RequestID && p.RecoverySecret == other.RecoverySecret
}

func (b LinkBinding) valid() bool {
	hash, hashErr := hex.DecodeString(b.KeyFingerprint)
	return validLinkResource(b.Resource) &&
		b.ClientID == "riido-cli" && linkBindingID(b.SessionID) && linkBindingID(b.ExpectedDeviceID) && hashErr == nil && len(hash) == 32 && hex.EncodeToString(hash) == b.KeyFingerprint
}

func validLinkResource(resource string) bool {
	u, err := url.Parse(resource)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && u.Path == "" && u.RawPath == "" && u.RawQuery == "" && !u.ForceQuery && u.Fragment == "" && resource == "https://"+u.Host
}

func linkBindingID(value string) bool {
	return value != "" && len(value) <= 512 && utf8.ValidString(value) && strings.TrimSpace(value) == value && !strings.ContainsAny(value, "/\\\x00\r\n\t")
}

func (p PendingLink) valid() bool {
	return p.LinkBinding.valid() && linkRandom(p.RequestID) && linkRandom(p.RecoverySecret) && (p.Phase == LinkPrepared || p.Phase == LinkCreated)
}

func linkRandom(value string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && len(decoded) == 32 && base64.RawURLEncoding.EncodeToString(decoded) == value
}

// BeginPendingLink holds the same resource lock as login, refresh and logout.
// Repeating a request returns its original pair, including after an ambiguous
// native save result. A different binding cannot replace an in-flight request.
func (s *ProtectedStore) BeginPendingLink(ctx context.Context, binding LinkBinding) (PendingLink, error) {
	unlock, record, err := s.lockLinkSession(ctx, binding)
	if err != nil {
		return PendingLink{}, err
	}
	defer unlock()
	if record.PendingLink != nil {
		return *record.PendingLink, nil
	}
	requestID, secret := make([]byte, 32), make([]byte, 32)
	if _, err = rand.Read(requestID); err != nil {
		return PendingLink{}, securecredential.ErrUnavailable
	}
	if _, err = rand.Read(secret); err != nil {
		return PendingLink{}, securecredential.ErrUnavailable
	}
	pending := PendingLink{LinkBinding: binding, Phase: LinkPrepared, RequestID: base64.RawURLEncoding.EncodeToString(requestID), RecoverySecret: base64.RawURLEncoding.EncodeToString(secret)}
	record.PendingLink = &pending
	if err = s.saveRecord(ctx, record); err != nil {
		return PendingLink{}, err
	}
	return pending, nil
}

// LoadPendingLink recovers the original binding from OS protected storage. A
// restarted CLI needs only the configured resource and current login session;
// receipt recovery does not require a live daemon or an in-memory device ID.
func (s *ProtectedStore) LoadPendingLink(ctx context.Context, expectedSessionID string) (*PendingLink, error) {
	unlock, record, err := s.lockProtectedSession(ctx, expectedSessionID)
	if err != nil {
		return nil, err
	}
	defer unlock()
	return record.PendingLink, nil
}

// ClearPendingLink cannot clear a newer request or save an old credential after
// logout or account replacement. The exact immutable request must match; phase
// progress does not replace request identity or establish server authority.
func (s *ProtectedStore) ClearPendingLink(ctx context.Context, pending PendingLink) error {
	if !pending.valid() {
		return securecredential.ErrUnavailable
	}
	unlock, record, err := s.lockLinkSession(ctx, pending.LinkBinding)
	if err != nil {
		return err
	}
	defer unlock()
	if record.PendingLink == nil {
		return nil
	}
	if !record.PendingLink.sameRequest(pending) {
		return securecredential.ErrUnavailable
	}
	record.PendingLink = nil
	return s.saveRecord(ctx, record)
}

func (s *ProtectedStore) lockLinkSession(ctx context.Context, binding LinkBinding) (func(), protectedSession, error) {
	if s == nil || !binding.valid() || binding.Resource != s.Resource {
		return nil, protectedSession{}, securecredential.ErrUnavailable
	}
	unlock, record, err := s.lockProtectedSession(ctx, binding.SessionID)
	if err != nil {
		return nil, protectedSession{}, err
	}
	if record.PendingLink != nil && record.PendingLink.LinkBinding != binding {
		unlock()
		return nil, protectedSession{}, securecredential.ErrUnavailable
	}
	return unlock, record, nil
}

func (s *ProtectedStore) lockProtectedSession(ctx context.Context, expectedSessionID string) (func(), protectedSession, error) {
	if s == nil || !validLinkResource(s.Resource) || !linkBindingID(expectedSessionID) {
		return nil, protectedSession{}, securecredential.ErrUnavailable
	}
	unlock, err := s.Lock(ctx)
	if err != nil {
		return nil, protectedSession{}, err
	}
	record, err := s.loadRecord(ctx)
	if err == nil && record.Credential.SessionID != expectedSessionID {
		err = userauth.ErrLoggedOut
	}
	if err != nil {
		unlock()
		return nil, protectedSession{}, err
	}
	return unlock, record, nil
}
