package userauth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

const desktopTokenFixture = `{"access_token":"desktop-access","refresh_token":"desktop-refresh","token_type":"Bearer","expires_in":600,"scope":"read","auth_origin":"desktop"}`

func desktopReply(t *testing.T, a *Authorization) []byte {
	t.Helper()
	r, err := a.DesktopRequest()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(desktopCallbackReply{Intent: fixtureIntent, Challenge: fixtureChallenge, ExpiresAt: time.Now().Add(10 * time.Minute).Unix(), KeyFingerprint: fixtureInstallationKey().KeyFingerprint, Code: fixtureChallenge, State: r.State, RedirectURL: r.RedirectURI + "?" + url.Values{"code": {fixtureChallenge}, "state": {r.State}}.Encode()})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func postDesktopReply(t *testing.T, a *Authorization, raw []byte, change func(*http.Request)) int {
	t.Helper()
	r, err := http.NewRequestWithContext(t.Context(), http.MethodPost, a.RedirectURI, bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/json")
	if change != nil {
		change(r)
	}
	response, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("callback cached")
	}
	return response.StatusCode
}

func TestDesktopAuthorizationBindsOriginalProofAndDirectOrigin(t *testing.T) {
	var form url.Values
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/oauth/token" || r.Header.Get("Authorization") != "" {
			t.Error("wrong token route or app token forwarded")
		}
		if r.ParseForm() != nil {
			t.Error("form")
		}
		form = r.PostForm
		fmt.Fprint(w, desktopTokenFixture)
	}))
	defer server.Close()
	c, _ := NewClient(configFor(server))
	var proof InstallationProof
	signed := 0
	a, err := c.BeginDesktopAuthorization(t.Context(), fixtureInstallationKey(), func(ctx context.Context, p InstallationProof) (string, error) {
		signed++
		proof = p
		if _, ok := ctx.Deadline(); !ok {
			t.Error("signer has no deadline")
		}
		return "MAYCAQECAQE", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	r, _ := a.DesktopRequest()
	raw, _ := json.Marshal(r)
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(raw, &fields)
	if len(fields) != 10 || strings.Contains(string(raw), a.verifier) || calls != 0 {
		t.Fatal("handoff leaked verifier or called API before app response")
	}
	if postDesktopReply(t, a, desktopReply(t, a), nil) != http.StatusOK {
		t.Fatal("callback")
	}
	if postDesktopReply(t, a, desktopReply(t, a), nil) != http.StatusConflict {
		t.Fatal("replay accepted")
	}
	credential, err := a.Wait(t.Context())
	if err != nil || credential.AccessToken != "desktop-access" || signed != 1 || calls != 1 {
		t.Fatal("Desktop flow failed", err, signed, calls)
	}
	sum := sha256.Sum256([]byte(form.Get("code_verifier")))
	if len(form) != 6 || form.Get("installation_signature") != "MAYCAQECAQE" || form.Get("redirect_uri") != r.RedirectURI || base64.RawURLEncoding.EncodeToString(sum[:]) != r.CodeChallenge || form.Get("code") != proof.Code {
		t.Fatal("exchange lost original PKCE or proof")
	}
	if proof.State != r.State || proof.CodeChallenge != r.CodeChallenge || proof.Resource != server.URL || proof.KeyFingerprint != fixtureInstallationKey().KeyFingerprint || proof.Intent != fixtureIntent || proof.ChallengeHash != "9f72ea0cf49536e3c66c787f705186df9a4378083753ae9536d65b3ad7fcddc4" {
		t.Fatal("proof changed request")
	}
	if _, err = a.Wait(t.Context()); err == nil {
		t.Fatal("flow reused")
	}
}

func TestDesktopAuthorizationDeniedAndCancelledNeverSigns(t *testing.T) {
	for _, code := range []string{"access_denied", "login_required", "temporarily_unavailable", "cancel"} {
		t.Run(code, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("terminal response reached API") }))
			defer server.Close()
			c, _ := NewClient(configFor(server))
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			a, err := c.BeginDesktopAuthorization(ctx, fixtureInstallationKey(), func(context.Context, InstallationProof) (string, error) {
				t.Error("terminal response signed")
				return "", nil
			})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			request, _ := a.DesktopRequest()
			if code == "cancel" {
				cancel()
			} else {
				raw, _ := json.Marshal(map[string]string{"state": request.State, "error": code})
				if postDesktopReply(t, a, raw, nil) != http.StatusOK {
					t.Fatal("denial rejected")
				}
			}
			if _, err = a.Wait(t.Context()); err == nil {
				t.Fatal("terminal response succeeded")
			}
		})
	}
}

func TestDesktopAuthorizationOriginMismatchOnlyCleansNewGrant(t *testing.T) {
	for _, body := range []string{strings.Replace(desktopTokenFixture, `"auth_origin":"desktop"`, `"auth_origin":"browser"`, 1), strings.Replace(desktopTokenFixture, `,"auth_origin":"desktop"`, "", 1), strings.Replace(desktopTokenFixture, `"auth_origin":"desktop"`, `"auth_origin":"desktop","auth_origin":"desktop"`, 1), strings.Replace(desktopTokenFixture, `"auth_origin"`, `"AUTH_ORIGIN"`, 1)} {
		t.Run(fmt.Sprint(len(body)), func(t *testing.T) {
			revokes := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/oauth/token" {
					fmt.Fprint(w, body)
					return
				}
				if r.URL.Path != "/oauth/revoke" || r.ParseForm() != nil || r.PostForm.Get("token") != "desktop-refresh" {
					t.Error("wrong cleanup target")
				}
				revokes++
			}))
			defer server.Close()
			c, _ := NewClient(configFor(server))
			a, err := c.BeginDesktopAuthorization(t.Context(), fixtureInstallationKey(), func(context.Context, InstallationProof) (string, error) { return "MAYCAQECAQE", nil })
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			if postDesktopReply(t, a, desktopReply(t, a), nil) != http.StatusOK {
				t.Fatal("callback")
			}
			var authErr *Error
			if credential, err := a.Wait(t.Context()); !errors.As(err, &authErr) || authErr.Code != "AUTH_ORIGIN_MISMATCH" || credential.AccessToken != "" || revokes != 1 {
				t.Fatal("unattested origin accepted", err, revokes)
			}
		})
	}
}

func TestDesktopLaunchURIContainsOnlyPublicRequestAndKnownApp(t *testing.T) {
	for resource, scheme := range map[string]string{"https://monolith.riido.io": "riido-cli-auth", "https://monolith.api.riido.io": "riido-staging-cli-auth"} {
		r := DesktopAuthorizationRequest{Resource: resource, State: fixtureIntent}
		address, err := r.LaunchURL()
		u, parseErr := url.Parse(address)
		if err != nil || parseErr != nil || u.Scheme != scheme || u.Host != "cli" || u.Path != "/authorize/v1" || len(u.Query()) != 1 {
			t.Fatal("wrong app URI", err)
		}
		raw, err := base64.RawURLEncoding.DecodeString(u.Query().Get("request"))
		var decoded DesktopAuthorizationRequest
		if err != nil || json.Unmarshal(raw, &decoded) != nil || decoded != r {
			t.Fatal("request changed")
		}
		if strings.Contains(fmt.Sprintf("%v %#v", r, r), fixtureIntent) {
			t.Fatal("request printed")
		}
	}
	for _, resource := range []string{"https://monolith.riido.io.evil.test", "https://monolith.riido.io@evil.test", "http://127.0.0.1:5432", "https://monolith.riido.io/"} {
		if address, err := (DesktopAuthorizationRequest{Resource: resource}).LaunchURL(); err == nil || address != "" {
			t.Fatal("unknown resource chose app")
		}
	}
}
