package standalonecli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestProtectedRefreshRecoveryRestartPreservesBothPendingOperations(t *testing.T) {
	store, native, credential, binding := newProtectedFixture(t)
	link, err := store.BeginPendingLink(t.Context(), binding)
	if err != nil {
		t.Fatal(err)
	}
	credential.ExpiresAt = time.Now()
	if err := saveProtectedCredential(t, store, credential); err != nil {
		t.Fatal(err)
	}
	first := protectedManager(t, store, func(request *http.Request) (*http.Response, error) {
		if request.ParseForm() != nil {
			t.Fatal("test form failed")
		}
		record, err := decodeProtectedSession(native.raw)
		if err != nil || record.Credential.RefreshRecoverySecret != request.Form.Get("recovery_secret") || record.Credential.RefreshToken != credential.RefreshToken || record.Credential.SessionID != credential.SessionID || record.PendingLink == nil || *record.PendingLink != link {
			t.Fatal("protected pending was not atomically saved before HTTP", err)
		}
		return nil, errors.New("synthetic committed rotation response lost")
	})
	if _, err := first.AccessToken(t.Context()); err == nil {
		t.Fatal("lost response reported success")
	}
	pending, err := decodeProtectedSession(native.raw)
	if err != nil || pending.Credential.RefreshRecoverySecret == "" {
		t.Fatal("pending missing", err)
	}
	for _, format := range []string{"%v", "%+v", "%#v"} {
		if strings.Contains(fmt.Sprintf(format, pending.Credential), pending.Credential.RefreshRecoverySecret) {
			t.Fatal("diagnostics disclosed recovery secret")
		}
	}
	restartedStore := &ProtectedStore{Native: native, Resource: store.Resource}
	restarted := protectedManager(t, restartedStore, func(request *http.Request) (*http.Response, error) {
		if request.ParseForm() != nil || request.Form.Get("recovery_secret") != pending.Credential.RefreshRecoverySecret || request.Form.Get("refresh_token") != credential.RefreshToken {
			t.Fatal("restart replaced original pair")
		}
		return protectedResponse(request, `{"access_token":"recovered","refresh_token":"recovered-refresh","token_type":"Bearer","expires_in":600,"refresh_recovery":"v1"}`), nil
	})
	if token, err := restarted.AccessToken(t.Context()); err != nil || token != "recovered" {
		t.Fatal("protected restart failed", err)
	}
	after, err := decodeProtectedSession(native.raw)
	if err != nil || after.Credential.RefreshRecoverySecret != "" || after.Credential.SessionID != credential.SessionID || after.PendingLink == nil || *after.PendingLink != link {
		t.Fatal("refresh receipt destroyed device-link pending or session", err)
	}
}

func TestProtectedRefreshRecoveryVersionAndMalformedInput(t *testing.T) {
	_, _, credential, _ := newProtectedFixture(t)
	validSecret := strings.Repeat("A", 43)
	for _, version := range []string{protectedSessionV1, protectedSessionV2} {
		legacy, _ := json.Marshal(map[string]any{"schema_version": version, "credential": credential, "pending_link": nil})
		got, err := decodeProtectedSession(legacy)
		if err != nil || got.Credential.SessionID != credential.SessionID || got.Credential.RefreshToken != credential.RefreshToken || !got.Credential.ExpiresAt.Equal(credential.ExpiresAt) {
			t.Fatal("legacy credential stopped loading", err)
		}
		withSecret := bytes.Replace(legacy, []byte(`"credential":{`), []byte(`"credential":{"refresh_recovery_secret":"`+validSecret+`",`), 1)
		if _, err := decodeProtectedSession(withSecret); err == nil {
			t.Fatal("legacy envelope smuggled a new pending secret")
		}
	}
	credential.RefreshRecoverySecret = validSecret
	raw, err := encodeCredential(credential)
	if err != nil {
		t.Fatal(err)
	}
	for _, malformed := range []string{`null`, `""`, `"short"`, `17`, `"` + strings.Repeat("A", 42) + `B"`} {
		bad := bytes.Replace(raw, []byte(`"refresh_recovery_secret":"`+validSecret+`"`), []byte(`"refresh_recovery_secret":`+malformed), 1)
		if _, err := decodeProtectedSession(bad); err == nil {
			t.Fatal("malformed recovery secret accepted")
		}
	}
	for _, mutate := range []func(*userauth.Credential){func(c *userauth.Credential) { c.SessionID = "" }, func(c *userauth.Credential) { c.RefreshToken = "" }} {
		bad := credential
		mutate(&bad)
		if _, err := encodeCredential(bad); err == nil {
			t.Fatal("unbound pending recovery persisted")
		}
	}
}
