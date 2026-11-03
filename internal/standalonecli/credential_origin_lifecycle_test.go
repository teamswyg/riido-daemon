package standalonecli

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/userauth"
)

// The HTTP server is a contract-shaped OAuth dependency; stores model the OS
// opaque boundary. This verifies local custody, not live server/OS acceptance.
func TestCredentialOriginLifecycleIsolation(t *testing.T) {
	for _, origin := range []string{"browser", "desktop"} {
		for _, revokeFailure := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/revoke-failure=%v", origin, revokeFailure), func(t *testing.T) {
				var revoked []string
				calls := 0
				server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.ParseForm() != nil {
						t.Error("invalid OAuth request")
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					if r.URL.Path == "/oauth/revoke" {
						revoked = append(revoked, r.Form.Get("token"))
						if revokeFailure && r.Form.Get("token") == "rotated-refresh" {
							w.WriteHeader(http.StatusServiceUnavailable)
						}
						return
					}
					if r.URL.Path != "/oauth/token" {
						t.Error("unexpected HTTP surface")
						w.WriteHeader(http.StatusNotImplemented)
						return
					}
					calls++
					if calls == 1 {
						conn, _, err := w.(http.Hijacker).Hijack()
						if err != nil {
							t.Error(err)
							return
						}
						_ = conn.Close()
						return
					}
					fmt.Fprint(w, `{"access_token":"rotated-access","refresh_token":"rotated-refresh","token_type":"Bearer","expires_in":600,"refresh_recovery":"v1"}`)
				}))
				defer server.Close()
				b, d, bn, dn := originStores(t)
				b.Resource, d.Resource = server.URL, server.URL
				own, other, ownNative, otherNative := b, d, bn, dn
				if origin == "desktop" {
					own, other, ownNative, otherNative = d, b, dn, bn
				}
				client, err := userauth.NewClient(userauth.Config{Resource: server.URL, ClientID: "riido-cli", HTTPClient: server.Client(), Endpoints: userauth.Endpoints{AuthorizeURL: server.URL + "/oauth/authorize", TokenURL: server.URL + "/oauth/token", RevokeURL: server.URL + "/oauth/revoke"}})
				if err != nil {
					t.Fatal(err)
				}
				for _, s := range []*ProtectedStore{own, other} {
					c := userauth.Credential{AccessToken: s.AuthOrigin + "-access", RefreshToken: s.AuthOrigin + "-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour), Binding: custodyBinding(s.AuthOrigin)}
					c.Binding.Resource = server.URL
					c.Binding.AccountID = s.AuthOrigin + "-account"
					if err := userauth.NewManager(client, s).Login(t.Context(), c); err != nil {
						t.Fatal(err)
					}
					if _, err := s.BeginPendingRepair(t.Context(), testRepairFingerprint()); err != nil {
						t.Fatal(err)
					}
				}
				otherBefore := bytes.Clone(otherNative.raw)
				repairBefore := bytes.Clone(other.RepairNative.(*protectedNativeFixture).raw)
				c, err := own.Load(t.Context())
				if err != nil {
					t.Fatal(err)
				}
				c.ExpiresAt = time.Now()
				if err = saveProtectedCredential(t, own, c); err != nil {
					t.Fatal(err)
				}
				if _, err = userauth.NewManager(client, own).AccessToken(t.Context()); err == nil {
					t.Fatal("response loss reported success")
				}
				pending, err := own.Load(t.Context())
				if err != nil || pending.RefreshRecoverySecret == "" || !userauth.SameCredentialBinding(pending.Binding, c.Binding) {
					t.Fatal("lost response discarded original custody", err)
				}
				if token, err := userauth.NewManager(client, own).AccessToken(t.Context()); err != nil || token != "rotated-access" {
					t.Fatal("restart recovery failed", err)
				}
				after, err := own.Load(t.Context())
				if err != nil || after.SessionID != c.SessionID || !userauth.SameCredentialBinding(after.Binding, c.Binding) || after.RefreshRecoverySecret != "" {
					t.Fatal("rotation changed custody", err)
				}
				replacement := after
				replacement.AccessToken, replacement.RefreshToken = "replacement-access", "replacement-refresh"
				binding := *after.Binding
				binding.AccountID = "replacement-account"
				replacement.Binding = &binding
				err = userauth.NewManager(client, own).Login(t.Context(), replacement)
				if revokeFailure && !errors.Is(err, userauth.ErrPreviousGrantRevocationFailed) || !revokeFailure && err != nil {
					t.Fatal("replacement revoke result changed", err)
				}
				current, err := own.Load(t.Context())
				if err != nil || current.SessionID == c.SessionID || current.AccessToken != "replacement-access" || current.Binding == nil || current.Binding.AccountID != "replacement-account" || !bytes.Equal(otherBefore, otherNative.raw) || !bytes.Equal(repairBefore, other.RepairNative.(*protectedNativeFixture).raw) {
					t.Fatal("replacement damaged own/new or opposite custody", err)
				}
				if err = userauth.NewManager(client, own).Logout(t.Context()); err != nil {
					t.Fatal(err)
				}
				if len(ownNative.raw) != 0 || len(own.RepairNative.(*protectedNativeFixture).raw) != 0 || !bytes.Equal(otherBefore, otherNative.raw) || !bytes.Equal(repairBefore, other.RepairNative.(*protectedNativeFixture).raw) {
					t.Fatal("logout crossed origin boundary")
				}
				if len(revoked) != 2 || revoked[0] != "rotated-refresh" || revoked[1] != "replacement-refresh" || calls != 2 {
					t.Fatal("unexpected revoke/refresh calls")
				}
				if strings.Contains(fmt.Sprint(c), "account") || strings.Contains(fmt.Sprintf("%#v", c.Binding), c.Binding.KeyFingerprint) {
					t.Fatal("debug formatting disclosed custody")
				}
			})
		}
	}
}
