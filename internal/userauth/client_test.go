package userauth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func configFor(server *httptest.Server) Config {
	return Config{ClientID: "riido-cli", Endpoints: Endpoints{AuthorizeURL: server.URL + "/oauth/authorize", TokenURL: server.URL + "/oauth/token", RevokeURL: server.URL + "/oauth/revoke"}, Resource: server.URL, Scopes: []string{"read"}}
}

func TestEndpointsRejectInsecureHosts(t *testing.T) {
	for _, address := range []string{"http://example.com", "http://localhost:1234", "http://127.0.0.1.evil.test", "https://user:secret@example.com", "https://example.com/#secret"} {
		cfg := Config{ClientID: "riido-cli", Endpoints: Endpoints{AuthorizeURL: address, TokenURL: address, RevokeURL: address}, Resource: address}
		if _, err := NewClient(cfg); err == nil {
			t.Errorf("accepted unsafe endpoint %q", address)
		}
	}
}

func TestAuthorizationPKCEAndState(t *testing.T) {
	var challenge string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
			t.Errorf("unexpected endpoint %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := r.ParseForm(); err != nil {
			t.Error(err)
		}
		verifier := r.Form.Get("code_verifier")
		digest := sha256.Sum256([]byte(verifier))
		if base64.RawURLEncoding.EncodeToString(digest[:]) != challenge {
			t.Error("PKCE verifier does not match challenge")
		}
		if r.Form.Get("client_id") != "riido-cli" || r.Form.Get("grant_type") != "authorization_code" || r.Form.Get("code") != "test-code" {
			t.Error("wrong exchange fields")
		}
		fmt.Fprint(w, `{"access_token":"test-access","refresh_token":"test-refresh","token_type":"Bearer","expires_in":3600,"scope":"read"}`)
	}))
	defer server.Close()
	client, err := NewClient(configFor(server))
	if err != nil {
		t.Fatal(err)
	}
	flow, err := client.BeginAuthorization(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer flow.Close()
	parsed, _ := url.Parse(flow.URL)
	q := parsed.Query()
	challenge = q.Get("code_challenge")
	if q.Get("code_challenge_method") != "S256" || q.Get("resource") != server.URL || len(q.Get("state")) < 32 {
		t.Fatal("missing PKCE/state/resource")
	}
	bad, err := http.Get(flow.RedirectURI + "?state=wrong&code=leaked")
	if err != nil {
		t.Fatal(err)
	}
	bad.Body.Close()
	if bad.StatusCode != http.StatusBadRequest {
		t.Fatalf("wrong state status=%d", bad.StatusCode)
	}
	good, err := http.Get(flow.RedirectURI + "?state=" + url.QueryEscape(q.Get("state")) + "&code=test-code")
	if err != nil {
		t.Fatal(err)
	}
	good.Body.Close()
	credential, err := flow.Wait(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if credential.AccessToken != "test-access" || credential.RefreshToken != "test-refresh" || time.Until(credential.ExpiresAt) < 59*time.Minute {
		t.Fatal("invalid credential")
	}
	if strings.Contains(fmt.Sprint(credential), "test-access") || strings.Contains(fmt.Sprintf("%#v", credential), "test-refresh") {
		t.Fatal("credential leaked through formatter")
	}
}

func TestOAuthResponseRedactionAndRedirects(t *testing.T) {
	leaked := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { leaked = true }))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("redirect") == "yes" {
			http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"invalid_grant","error_description":"test-secret-must-not-leak"}`)
	}))
	defer server.Close()
	cfg := configFor(server)
	client, _ := NewClient(cfg)
	_, err := client.Refresh(context.Background(), "refresh-secret")
	if err == nil || strings.Contains(err.Error(), "test-secret") || strings.Contains(err.Error(), "refresh-secret") {
		t.Fatalf("unsafe error: %v", err)
	}
	cfg.Endpoints.TokenURL += "?redirect=yes"
	client, err = NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Refresh(context.Background(), "refresh-secret")
	if err == nil || leaked {
		t.Fatal("credential-bearing redirect followed")
	}
}

func TestAuthorizationCancellationClosesListener(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	ctx, cancel := context.WithCancel(context.Background())
	flow, err := client.BeginAuthorization(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	_, err = flow.Wait(ctx)
	if err == nil {
		t.Fatal("canceled authorization succeeded")
	}
	response, err := http.Get(flow.RedirectURI)
	if err == nil {
		_, _ = io.Copy(io.Discard, response.Body)
		response.Body.Close()
		t.Fatal("listener remains open")
	}
}
