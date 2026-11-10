package standalonecli

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
)

// A transfer survives logout so the original key and fresh human consent can
// reconcile its receipt. No bearer token is stored here. The shared operation
// lock serializes every reader/writer; the short credential lock fences login.
type transferJournal struct {
	Schema        string `json:"schema"`
	Resource      string `json:"resource"`
	Origin        string `json:"origin"`
	Session       string `json:"session"`
	Fingerprint   string `json:"fingerprint"`
	Device        string `json:"device"`
	Request       string `json:"request"`
	Secret        string `json:"secret"`
	Created       bool   `json:"created"`
	RepairRequest string `json:"repair_request"`
	RepairSecret  string `json:"repair_secret"`
	RepairCreated bool   `json:"repair_created"`
}

func (transferJournal) String() string     { return "[redacted transfer journal]" }
func (j transferJournal) GoString() string { return j.String() }
func newTransferPair() (string, string, error) {
	var bytes [64]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes[:32]), base64.RawURLEncoding.EncodeToString(bytes[32:]), nil
}

func (s *ProtectedStore) validateTransfer(j transferJournal) bool {
	binding := LinkBinding{Resource: j.Resource, ClientID: "riido-cli", SessionID: j.Session, ExpectedDeviceID: j.Device, KeyFingerprint: j.Fingerprint}
	hasRepair := j.RepairRequest != "" || j.RepairSecret != "" || j.RepairCreated
	return s != nil && s.TransferNative != nil && j.Schema == "riido-cli-transfer.v1" && j.Origin == s.authOrigin() && j.Resource == s.Resource && binding.valid() && linkRandom(j.Request) && linkRandom(j.Secret) && (!hasRepair || j.Created && linkRandom(j.RepairRequest) && linkRandom(j.RepairSecret) && j.RepairRequest != j.Request)
}

func (s *ProtectedStore) loadTransfer(ctx context.Context) (*transferJournal, error) {
	if s.TransferNative == nil {
		return nil, securecredential.ErrUnavailable
	}
	body, err := s.TransferNative.Load(ctx)
	if errors.Is(err, securecredential.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, mapStorageError(err)
	}
	if len(body) > 4096 {
		return nil, securecredential.ErrUnavailable
	}
	fields, err := protectedObject(body)
	if err != nil || len(fields) != 12 || !protectedFields(fields, "schema", "resource", "origin", "session", "fingerprint", "device", "request", "secret", "created", "repair_request", "repair_secret", "repair_created") {
		return nil, securecredential.ErrUnavailable
	}
	for _, v := range fields {
		if string(v) == "null" {
			return nil, securecredential.ErrUnavailable
		}
	}
	var j transferJournal
	if json.Unmarshal(body, &j) != nil || !s.validateTransfer(j) {
		return nil, securecredential.ErrUnavailable
	}
	return &j, nil
}

func (s *ProtectedStore) saveTransfer(ctx context.Context, j transferJournal) error {
	if !s.validateTransfer(j) {
		return securecredential.ErrUnavailable
	}
	body, err := json.Marshal(j)
	if err != nil {
		return securecredential.ErrUnavailable
	}
	return mapStorageError(s.TransferNative.Save(ctx, body))
}
