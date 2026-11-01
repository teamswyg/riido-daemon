package standalonecli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type desktopLoginNative struct {
	*protectedNativeFixture
	confirmErr, errorSign error
	confirms, signs       int
}

func (n *desktopLoginNative) EnsureInstallationKey(context.Context, string) (securecredential.InstallationPublicKey, error) {
	n.confirms++
	k := testInstallationKey()
	return securecredential.InstallationPublicKey{Algorithm: k.Algorithm, PublicKey: k.PublicKey, KeyFingerprint: k.KeyFingerprint}, n.confirmErr
}

func (n *desktopLoginNative) SignInstallationProof(context.Context, securecredential.InstallationPublicKey, securecredential.InstallationProof) (string, error) {
	n.signs++
	return "MAYCAQECAQE", n.errorSign
}

func desktopLoginCallback(ctx context.Context, request userauth.DesktopAuthorizationRequest, denied bool) error {
	code := "IiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiIiI"
	reply := map[string]any{"installation_intent": "ERERERERERERERERERERERERERERERERERERERERERE", "challenge": code, "expires_at": time.Now().Add(10 * time.Minute).Unix(), "key_fingerprint": testInstallationKey().KeyFingerprint, "code": code, "state": request.State, "redirect_url": request.RedirectURI + "?" + url.Values{"code": {code}, "state": {request.State}}.Encode()}
	if denied {
		reply = map[string]any{"state": request.State, "error": "access_denied"}
	}
	raw, err := json.Marshal(reply)
	if err != nil {
		return err
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, request.RedirectURI, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	r.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return errors.New("callback failed")
	}
	return nil
}

func TestDesktopLoginProtectedCustodyPreservesBrowser(t *testing.T) {
	for _, scenario := range []string{"success", "wrong-origin", "denied", "presence-refusal", "signer-refusal", "identity-failure", "old-revoke-failure", "app-unavailable"} {
		t.Run(scenario, func(t *testing.T) {
			var revoked []string
			var requests []string
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests = append(requests, r.URL.Path)
				switch r.URL.Path {
				case "/oauth/token":
					origin := "desktop"
					if scenario == "wrong-origin" {
						origin = "browser"
					}
					_, _ = fmt.Fprintf(w, `{"access_token":"new-desktop-access","refresh_token":"new-desktop-refresh","token_type":"Bearer","expires_in":600,"scope":"read","auth_origin":%q}`, origin)
				case "/graphql":
					if scenario == "identity-failure" {
						w.WriteHeader(http.StatusServiceUnavailable)
						return
					}
					fmt.Fprint(w, `{"data":{"getMyAccountInfo":{"id":"new-desktop-user"}}}`)
				case "/oauth/revoke":
					if r.ParseForm() != nil {
						t.Error("revoke form")
					}
					revoked = append(revoked, r.PostForm.Get("token"))
					if scenario == "old-revoke-failure" {
						w.WriteHeader(http.StatusServiceUnavailable)
					}
				default:
					t.Error("unexpected API", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			browser, desktop, bn, dn := originStores(t)
			browser.Resource, desktop.Resource = server.URL, server.URL
			for _, s := range []*ProtectedStore{browser, desktop} {
				previous := userauth.Credential{SessionID: "previous-" + s.AuthOrigin, AccessToken: s.AuthOrigin + "-old-access", RefreshToken: s.AuthOrigin + "-old-refresh", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}
				previous.Binding = custodyBinding(s.AuthOrigin)
				previous.Binding.Resource = server.URL
				previous.Binding.AccountID = "previous-" + s.AuthOrigin
				previous.Binding.KeyFingerprint = testInstallationKey().KeyFingerprint
				if err := saveProtectedCredential(t, s, previous); err != nil {
					t.Fatal(err)
				}
			}
			browserBefore, desktopBefore := bytes.Clone(bn.raw), bytes.Clone(dn.raw)
			native := &desktopLoginNative{protectedNativeFixture: dn}
			desktop.Native = native
			if scenario == "presence-refusal" {
				native.confirmErr = userauth.ErrStoreLocked
			}
			if scenario == "signer-refusal" {
				native.errorSign = userauth.ErrStoreLocked
			}
			opened := 0
			service, err := New(Config{APIOrigin: server.URL, AuthOrigin: "desktop", HTTPClient: server.Client(), Store: desktop, OpenBrowser: func(context.Context, string) error { t.Error("Desktop fell back to browser"); return nil }, OpenDesktop: func(ctx context.Context, r userauth.DesktopAuthorizationRequest) error {
				opened++
				if scenario == "app-unavailable" {
					return ErrDesktopUnavailable
				}
				return desktopLoginCallback(ctx, r, scenario == "denied")
			}})
			if err != nil {
				t.Fatal(err)
			}
			result, err := service.Login(t.Context())
			if !bytes.Equal(browserBefore, bn.raw) {
				t.Fatal("Desktop changed browser credential")
			}
			success := scenario == "success" || scenario == "old-revoke-failure"
			if !success {
				if err == nil || result.Authenticated || !bytes.Equal(desktopBefore, dn.raw) {
					t.Fatal("failed login replaced old session", err)
				}
				if scenario == "wrong-origin" || scenario == "identity-failure" {
					if len(revoked) != 1 || revoked[0] != "new-desktop-refresh" {
						t.Fatal("wrong cleanup")
					}
				} else if len(requests) != 0 {
					t.Fatal("rejected login reached API")
				}
				return
			}
			if err != nil || !result.Authenticated || result.AuthOrigin != "desktop" || result.Identity.ID != "new-desktop-user" || native.confirms != 1 || native.signs != 1 || opened != 1 {
				t.Fatal("Desktop login failed", err)
			}
			credential, err := desktop.Load(t.Context())
			if err != nil || credential.Binding == nil || !credential.Binding.Valid() || credential.Binding.AuthOrigin != "desktop" || credential.Binding.AccountID != "new-desktop-user" || credential.Binding.KeyFingerprint != testInstallationKey().KeyFingerprint {
				t.Fatal("saved binding differs", err)
			}
			if len(revoked) != 1 || revoked[0] != "desktop-old-refresh" {
				t.Fatal("other origin revoked")
			}
			if scenario == "old-revoke-failure" && len(result.Warnings) != 1 {
				t.Fatal("partial cleanup hidden")
			}
		})
	}
}
