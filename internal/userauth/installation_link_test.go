package userauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func linkFixtureRequest() DeviceLinkRequest {
	return DeviceLinkRequest{RequestID: base64.RawURLEncoding.EncodeToString(make([]byte, 32)), RecoverySecret: base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("s", 32))), ExpectedDeviceID: "original-device"}
}
func linkFixtureClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	cfg := configFor(server)
	cfg.ClientID = "riido-cli"
	cfg.Scopes = []string{"read"}
	cfg.HTTPClient = server.Client()
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse(server.URL)
	jar.SetCookies(u, []*http.Cookie{{Name: "ambient", Value: "must-not-send"}})
	cfg.HTTPClient.Jar = jar
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
func linkFixtureStore() *fakeStore {
	credential := validCredential()
	credential.Scope = "read"
	return newFakeStore(credential)
}
func linkRecoveryInput(r DeviceLinkRequest) DeviceLinkRecoveryRequest {
	return DeviceLinkRecoveryRequest{RequestID: r.RequestID, RecoverySecret: r.RecoverySecret, ExpectedDeviceID: r.ExpectedDeviceID, KeyFingerprint: strings.Repeat("a", 64)}
}
func linkReply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func TestDeviceLinkCreateUsesFixedAuthenticatedPOSTAndPreservesOriginalExpiry(t *testing.T) {
	input := linkFixtureRequest()
	calls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "POST" || r.URL.Path != "/oauth/cli/installation-links" || r.URL.RawQuery != "" || r.Header.Get("Authorization") != "Bearer old-access" || r.Header.Get("Cookie") != "" || r.Header.Get("Origin") != "" {
			t.Error("wrong create boundary")
		}
		var fields map[string]string
		if json.NewDecoder(r.Body).Decode(&fields) != nil || len(fields) != 3 || fields["request_id"] != input.RequestID || fields["recovery_secret"] != input.RecoverySecret || fields["expected_device_id"] != input.ExpectedDeviceID {
			t.Error("changed original pair or public authority")
		}
		linkReply(w, 201, DeviceLinkApproval{RequestID: input.RequestID, State: "approved", ExpiresAt: 1})
	}))
	defer server.Close()
	store := linkFixtureStore()
	manager := NewManager(linkFixtureClient(t, server), store)
	result, err := manager.CreateDeviceLink(context.Background(), "fixture-session", input)
	if err != nil || result.RequestID != input.RequestID || result.State != "approved" || result.ExpiresAt != 1 || calls != 1 {
		t.Fatal("original create response was lost or renewed", err)
	}
	if !store.present || store.credential.AccessToken != "old-access" {
		t.Fatal("link create modified credentials")
	}
}

func TestDeviceLinkRecoveryUsesOnlyOriginalSecretWithoutOAuth(t *testing.T) {
	input := linkFixtureRequest()
	calls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "POST" || r.URL.Path != "/oauth/cli/installation-links/recover" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" || r.Header.Get("Origin") != "" {
			t.Error("recovery acquired ambient authority")
		}
		var fields map[string]string
		if json.NewDecoder(r.Body).Decode(&fields) != nil || len(fields) != 2 || fields["request_id"] != input.RequestID || fields["recovery_secret"] != input.RecoverySecret {
			t.Error("recovery replaced pair or chose authority")
		}
		linkReply(w, 200, map[string]string{"request_id": input.RequestID, "state": "linked"})
	}))
	defer server.Close()
	result, err := linkFixtureClient(t, server).RecoverDeviceLink(context.Background(), linkRecoveryInput(input))
	if err != nil || result.State != "linked" || result.RequestID != input.RequestID || result.Proof != nil || calls != 1 {
		t.Fatal("secret-only recovery failed", err)
	}
}

func TestDeviceLinkCreateRejectsSessionReplacementBeforeAndDuringHTTP(t *testing.T) {
	input := linkFixtureRequest()
	store := linkFixtureStore()
	calls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		unlock, err := store.Lock(r.Context())
		if err != nil {
			t.Error(err)
			return
		}
		replacement := validCredential()
		replacement.SessionID = "replacement-session"
		replacement.AccessToken = "replacement-access"
		replacement.Scope = "read"
		_ = store.Save(r.Context(), replacement)
		unlock()
		linkReply(w, 201, DeviceLinkApproval{RequestID: input.RequestID, State: "pending", ExpiresAt: 1})
	}))
	defer server.Close()
	manager := NewManager(linkFixtureClient(t, server), store)
	if _, err := manager.CreateDeviceLink(context.Background(), "old-session", input); !errors.Is(err, ErrLoggedOut) || calls != 0 {
		t.Fatal("wrong original session reached server", err)
	}
	if _, err := manager.CreateDeviceLink(context.Background(), "fixture-session", input); !errors.Is(err, ErrLoggedOut) || calls != 1 {
		t.Fatal("replacement response attached old request to new session", err)
	}
	if store.credential.SessionID != "replacement-session" {
		t.Fatal("response rewrote replacement session")
	}
}

func TestDeviceLinkRecoveryErrorsNeverBecomeOAuthReplay(t *testing.T) {
	for _, status := range []int{400, 401, 403, 408, 429, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				linkReply(w, status, map[string]string{"error": "invalid_link_recovery", "error_description": "secret-provider-body"})
			}))
			defer server.Close()
			_, err := linkFixtureClient(t, server).RecoverDeviceLink(context.Background(), linkRecoveryInput(linkFixtureRequest()))
			var linkErr *DeviceLinkError
			if !errors.As(err, &linkErr) || linkErr.StatusCode != status || errors.Is(err, ErrReauthRequired) || strings.Contains(err.Error(), "secret-provider-body") {
				t.Fatal("link failure lost sanitized boundary", err)
			}
			if linkErr.Retryable != (status == 408 || status == 429 || status >= 500) {
				t.Fatal("wrong retry classification")
			}
		})
	}
}

func TestDeviceLinkRecoveryRejectsMalformedEnvelope(t *testing.T) {
	input := linkFixtureRequest()
	for _, body := range []string{`{"request_id":"` + input.RequestID + `","state":"linked","proof":null}`, `{"request_id":"` + input.RequestID + `","state":"linked","state":"proof_required"}`, `{"request_id":"other","state":"linked"}`, `{"request_id":"` + input.RequestID + `","state":"unknown"}`, `{"request_id":"` + input.RequestID + `","state":"linked"} {}`, strings.Repeat("x", 4097)} {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, body)
		}))
		_, err := linkFixtureClient(t, server).RecoverDeviceLink(context.Background(), linkRecoveryInput(input))
		server.Close()
		var linkErr *DeviceLinkError
		if !errors.As(err, &linkErr) || linkErr.Code != "LINK_INVALID_RESPONSE" {
			t.Fatal("malformed response not rejected", err)
		}
	}
}
