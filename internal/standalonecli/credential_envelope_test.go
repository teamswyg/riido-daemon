package standalonecli

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestNativeStorageCodecReadsVersionedCredentialEnvelope(t *testing.T) {
	credential := userauth.Credential{SessionID: "preserved-session", AccessToken: "synthetic-access", RefreshToken: "synthetic-refresh", TokenType: "Bearer", ExpiresAt: time.Date(2026, 10, 29, 5, 0, 0, 0, time.UTC)}
	raw, err := json.Marshal(map[string]any{"schema_version": "riido-cli-protected-session.v1", "credential": credential, "pending_link": nil})
	if err != nil {
		t.Fatal(err)
	}
	got, err := decodeCredential(raw)
	if err != nil || !reflect.DeepEqual(got, credential) {
		t.Fatal("versioned protected record did not preserve the current credential", err)
	}
}

func TestNativeStorageCodecRejectsDuplicateProtectedKeys(t *testing.T) {
	credential := userauth.Credential{SessionID: "preserved-session", AccessToken: "synthetic-access", TokenType: "Bearer"}
	legacy, err := json.Marshal(credential)
	if err != nil {
		t.Fatal(err)
	}
	duplicateCredential := `{"session_id":"other-session",` + string(legacy[1:])
	pending := PendingLink{LinkBinding: LinkBinding{Resource: protectedTestResource, ClientID: "riido-cli", SessionID: credential.SessionID, KeyFingerprint: strings.Repeat("a", 64), ExpectedDeviceID: "device-original"}, RequestID: strings.Repeat("A", 43), RecoverySecret: strings.Repeat("C", 42) + "A"}
	pendingRaw, err := json.Marshal(pending)
	if err != nil {
		t.Fatal(err)
	}
	duplicatePending := `{"request_id":"` + strings.Repeat("B", 42) + `A",` + string(pendingRaw[1:])
	for name, raw := range map[string]string{
		"schema":            `{"schema_version":"unknown-future","schema_version":"riido-cli-protected-session.v1","credential":` + string(legacy) + `,"pending_link":null}`,
		"schema-case-alias": `{"schema_version":"unknown-future","Schema_Version":"riido-cli-protected-session.v1","credential":` + string(legacy) + `,"pending_link":null}`,
		"legacy":            duplicateCredential,
		"credential":        `{"schema_version":"riido-cli-protected-session.v1","credential":` + duplicateCredential + `,"pending_link":null}`,
		"pending":           `{"schema_version":"riido-cli-protected-session.v1","credential":` + string(legacy) + `,"pending_link":` + duplicatePending + `}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, decodeErr := decodeCredential([]byte(raw)); decodeErr == nil {
				t.Fatal("ambiguous protected JSON was accepted")
			}
		})
	}
}

func TestNativeStorageCodecMigratesLegacyOnlyOnSave(t *testing.T) {
	credential := userauth.Credential{SessionID: "legacy-session", AccessToken: "synthetic-access", RefreshToken: "synthetic-refresh", TokenType: "Bearer", ExpiresAt: time.Date(2026, 10, 29, 5, 0, 0, 0, time.UTC)}
	legacy, err := json.Marshal(credential)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := decodeCredential(legacy)
	if err != nil || !reflect.DeepEqual(loaded, credential) {
		t.Fatal("legacy credential became unreadable", err)
	}
	saved, err := encodeCredential(loaded)
	var envelope struct {
		SchemaVersion string              `json:"schema_version"`
		Credential    userauth.Credential `json:"credential"`
	}
	if err != nil || json.Unmarshal(saved, &envelope) != nil || envelope.SchemaVersion != protectedSessionVersion || !reflect.DeepEqual(envelope.Credential, credential) {
		t.Fatal("next credential save did not migrate the original credential")
	}
}
