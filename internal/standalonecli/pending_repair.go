package standalonecli

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

const protectedRepairVersion = "riido-cli-protected-repair.v1"

type protectedRepairNativeStore interface {
	Load(context.Context) ([]byte, error)
	Save(context.Context, []byte) error
	Delete(context.Context) error
}

// PendingRepair is separate from OAuth custody but fenced by the same lock and
// captured login identity. The empty SessionID deliberately means logged out.
type PendingRepair struct {
	SchemaVersion  string    `json:"schema_version"`
	Resource       string    `json:"resource"`
	ClientID       string    `json:"client_id"`
	SessionID      string    `json:"session_id"`
	KeyFingerprint string    `json:"key_fingerprint"`
	RequestID      string    `json:"request_id"`
	RecoverySecret string    `json:"recovery_secret"`
	Phase          LinkPhase `json:"phase"`
}

func (PendingRepair) String() string     { return "[redacted protected repair request]" }
func (p PendingRepair) GoString() string { return p.String() }
func (p PendingRepair) valid() bool {
	digest, err := hex.DecodeString(p.KeyFingerprint)
	return p.SchemaVersion == protectedRepairVersion && validLinkResource(p.Resource) && p.ClientID == "riido-cli" && (p.SessionID == "" || linkBindingID(p.SessionID)) && err == nil && len(digest) == 32 && hex.EncodeToString(digest) == p.KeyFingerprint && linkRandom(p.RequestID) && linkRandom(p.RecoverySecret) && (p.Phase == LinkPrepared || p.Phase == LinkCreated)
}

func (p PendingRepair) sameRequest(other PendingRepair) bool {
	p.Phase = LinkPrepared
	other.Phase = LinkPrepared
	return p == other
}

func (s *ProtectedStore) BeginPendingRepair(ctx context.Context, fingerprint string) (PendingRepair, error) {
	unlock, sessionID, err := s.lockRepair(ctx)
	if err != nil {
		return PendingRepair{}, err
	}
	defer unlock()
	current, err := s.loadRepair(ctx, sessionID)
	if err != nil {
		return PendingRepair{}, err
	}
	if current != nil {
		if current.KeyFingerprint != fingerprint {
			return PendingRepair{}, repairLocalError("REPAIR_BINDING_MISMATCH")
		}
		return *current, nil
	}
	id, secret := make([]byte, 32), make([]byte, 32)
	if _, err = rand.Read(id); err != nil {
		return PendingRepair{}, securecredential.ErrUnavailable
	}
	if _, err = rand.Read(secret); err != nil {
		return PendingRepair{}, securecredential.ErrUnavailable
	}
	p := PendingRepair{SchemaVersion: protectedRepairVersion, Resource: s.Resource, ClientID: "riido-cli", SessionID: sessionID, KeyFingerprint: fingerprint, RequestID: base64.RawURLEncoding.EncodeToString(id), RecoverySecret: base64.RawURLEncoding.EncodeToString(secret), Phase: LinkPrepared}
	if err = s.saveRepair(ctx, p); err != nil {
		return PendingRepair{}, err
	}
	return p, nil
}

func (s *ProtectedStore) LoadPendingRepair(ctx context.Context) (*PendingRepair, error) {
	unlock, sessionID, err := s.lockRepair(ctx)
	if err != nil {
		return nil, err
	}
	defer unlock()
	return s.loadRepair(ctx, sessionID)
}

func (s *ProtectedStore) MarkPendingRepairCreated(ctx context.Context, p PendingRepair) (PendingRepair, error) {
	unlock, current, err := s.lockExactRepair(ctx, p)
	if err != nil {
		return PendingRepair{}, err
	}
	defer unlock()
	if current.Phase == LinkCreated {
		return current, nil
	}
	current.Phase = LinkCreated
	if err = s.saveRepair(ctx, current); err != nil {
		return PendingRepair{}, err
	}
	return current, nil
}

func (s *ProtectedStore) ClearPendingRepair(ctx context.Context, p PendingRepair) error {
	unlock, _, err := s.lockExactRepair(ctx, p)
	if err != nil {
		return err
	}
	defer unlock()
	return mapStorageError(s.RepairNative.Delete(ctx))
}

func (s *ProtectedStore) lockExactRepair(ctx context.Context, p PendingRepair) (func(), PendingRepair, error) {
	if !p.valid() {
		return nil, PendingRepair{}, securecredential.ErrUnavailable
	}
	unlock, sessionID, err := s.lockRepair(ctx)
	if err != nil {
		return nil, PendingRepair{}, err
	}
	current, err := s.loadRepair(ctx, sessionID)
	if err == nil && (current == nil || !current.sameRequest(p)) {
		err = repairLocalError("REPAIR_REQUEST_CHANGED")
	}
	if err != nil {
		unlock()
		return nil, PendingRepair{}, err
	}
	return unlock, *current, nil
}

func (s *ProtectedStore) lockRepair(ctx context.Context) (func(), string, error) {
	if s == nil || s.Native == nil || s.RepairNative == nil || !validLinkResource(s.Resource) {
		return nil, "", securecredential.ErrUnavailable
	}
	unlock, err := s.Lock(ctx)
	if err != nil {
		return nil, "", err
	}
	credential, err := s.Load(ctx)
	if errors.Is(err, userauth.ErrNotAuthenticated) {
		return unlock, "", nil
	}
	if err == nil && !linkBindingID(credential.SessionID) {
		err = securecredential.ErrUnavailable
	}
	if err != nil {
		unlock()
		return nil, "", err
	}
	return unlock, credential.SessionID, nil
}

func (s *ProtectedStore) loadRepair(ctx context.Context, sessionID string) (*PendingRepair, error) {
	raw, err := s.RepairNative.Load(ctx)
	if errors.Is(err, securecredential.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, mapStorageError(err)
	}
	if len(raw) == 0 || len(raw) > 4096 {
		return nil, securecredential.ErrUnavailable
	}
	fields, err := protectedObject(raw)
	if err != nil || len(fields) != 8 || !protectedFields(fields, "schema_version", "resource", "client_id", "session_id", "key_fingerprint", "request_id", "recovery_secret", "phase") {
		return nil, securecredential.ErrUnavailable
	}
	// Null cannot stand in for the explicit empty no-session binding.
	for _, raw := range fields {
		var value string
		if string(raw) == "null" || json.Unmarshal(raw, &value) != nil {
			return nil, securecredential.ErrUnavailable
		}
	}
	var p PendingRepair
	if json.Unmarshal(raw, &p) != nil || !p.valid() || p.Resource != s.Resource {
		return nil, securecredential.ErrUnavailable
	}
	if p.SessionID != sessionID {
		return nil, userauth.ErrLoggedOut
	}
	return &p, nil
}

func (s *ProtectedStore) saveRepair(ctx context.Context, p PendingRepair) error {
	if !p.valid() {
		return securecredential.ErrUnavailable
	}
	raw, err := json.Marshal(p)
	if err != nil || len(raw) > 4096 {
		return securecredential.ErrUnavailable
	}
	return mapStorageError(s.RepairNative.Save(ctx, raw))
}

// ClearLocalRequestsOnLogout is called only while the authentication manager
// holds the credential lock. Keys remain, but an old in-flight request stops.
func (s *ProtectedStore) ClearLocalRequestsOnLogout(ctx context.Context) error {
	if s.RepairNative == nil {
		return nil
	}
	err := s.RepairNative.Delete(ctx)
	if errors.Is(err, securecredential.ErrNotFound) {
		return nil
	}
	return mapStorageError(err)
}
func repairLocalError(code string) error { return &userauth.InstallationRepairError{Code: code} }
