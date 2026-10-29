package securecredential

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
)

func TestExistingInstallationKeyUsesOnlyExistingTypedOperation(t *testing.T) {
	key, proof, _ := installationLiveFixture(t)
	calls := 0
	s := testStore(func(_ context.Context, _ string, raw []byte) ([]byte, error) {
		calls++
		var sent request
		if json.Unmarshal(raw, &sent) != nil || sent.Operation != "installation-existing-key" || !sent.Interactive || sent.Key != installationStorageKey(proof.Resource) {
			t.Fatal("existing key lookup changed its protected scope or operation")
		}
		data, err := base64.StdEncoding.DecodeString(sent.Value)
		if err != nil {
			t.Fatal(err)
		}
		var scope map[string]string
		if json.Unmarshal(data, &scope) != nil || len(scope) != 2 || scope["resource"] != proof.Resource || scope["client_id"] != "riido-cli" {
			t.Fatal("existing lookup accepted arbitrary key fields")
		}
		return json.Marshal(map[string]any{"version": 1, "ok": true, "data": key})
	})
	got, err := s.ExistingInstallationKey(t.Context(), proof.Resource)
	if err != nil || got != key || calls != 1 {
		t.Fatal("existing protected public key unavailable", err)
	}
	if s.Key != "account-key-0123456789" {
		t.Fatal("lookup changed credential storage identity")
	}
}

func TestExistingInstallationKeyMissingAndOldHelperNeverEnsure(t *testing.T) {
	_, proof, _ := installationLiveFixture(t)
	for _, code := range []string{"NOT_FOUND", "INVALID_REQUEST", "UNAVAILABLE"} {
		t.Run(code, func(t *testing.T) {
			calls := 0
			s := testStore(func(_ context.Context, _ string, raw []byte) ([]byte, error) {
				calls++
				var sent request
				if json.Unmarshal(raw, &sent) != nil || sent.Operation != "installation-existing-key" {
					t.Fatal("missing lookup tried to ensure/create a key")
				}
				return json.Marshal(map[string]any{"version": 1, "ok": false, "error": map[string]string{"code": code, "message": "synthetic-private-detail"}})
			})
			_, err := s.ExistingInstallationKey(t.Context(), proof.Resource)
			if err == nil || calls != 1 || code == "NOT_FOUND" && !errors.Is(err, ErrNotFound) || code != "NOT_FOUND" && !errors.Is(err, ErrUnavailable) {
				t.Fatal("existing key refusal lost its meaning", err)
			}
		})
	}
}

func TestExistingInstallationKeyRequiresInteractiveVerifiedHelper(t *testing.T) {
	_, proof, _ := installationLiveFixture(t)
	calls := 0
	s := testStore(func(context.Context, string, []byte) ([]byte, error) { calls++; return nil, errors.New("must not run") })
	s.Interactive = false
	if _, err := s.ExistingInstallationKey(t.Context(), proof.Resource); !errors.Is(err, ErrInteractionRequired) || calls != 0 {
		t.Fatal("noninteractive lookup reached key store", err)
	}
	s.Interactive = true
	s.verify = func(context.Context, string) error { return ErrUnavailable }
	if _, err := s.ExistingInstallationKey(t.Context(), proof.Resource); !errors.Is(err, ErrUnavailable) || calls != 0 {
		t.Fatal("unverified helper received request", err)
	}
}
