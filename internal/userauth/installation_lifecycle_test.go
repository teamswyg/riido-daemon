package userauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

func TestInstallationAdmissionCancellationClosesListener(t *testing.T) {
	redirect := make(chan string, 1)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Redirect string `json:"redirect_uri"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Error(err)
		}
		redirect <- input.Redirect
		<-release
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		flow, err := client.BeginInstallationAuthorization(ctx, fixtureInstallationKey(), func(context.Context, InstallationProof) (string, error) {
			t.Error("cancelled admission signed")
			return "", nil
		})
		if flow != nil {
			flow.Close()
		}
		done <- err
	}()
	address := <-redirect
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("cancellation=%v", err)
		}
	case <-time.After(time.Second):
		close(release)
		t.Fatal("admission cancellation did not finish")
	}
	close(release)
	assertInstallationListenerClosed(t, address)
}

func TestInstallationSignerCancellationStopsBeforeTokenAndClosesListener(t *testing.T) {
	key := fixtureInstallationKey()
	var tokens atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/cli/installation-intents" {
			writeFixtureIntent(w, key, time.Now().Add(time.Minute).Unix())
			return
		}
		tokens.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	entered := make(chan struct{})
	flow, err := client.BeginInstallationAuthorization(ctx, key, func(signCtx context.Context, _ InstallationProof) (string, error) {
		close(entered)
		<-signCtx.Done()
		return "MAYCAQECAQE", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer flow.Close()
	installationCallback(t, flow, "fixture-code")
	done := make(chan error, 1)
	go func() { _, err := flow.Wait(context.Background()); done <- err }()
	<-entered
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("signer cancellation=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("signer did not receive cancellation")
	}
	if tokens.Load() != 0 {
		t.Fatal("cancelled signer reached token endpoint")
	}
	assertInstallationListenerClosed(t, flow.RedirectURI)
}

func TestExpiredInstallationNeverSignsOrExchanges(t *testing.T) {
	key := fixtureInstallationKey()
	tokens := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/cli/installation-intents" {
			writeFixtureIntent(w, key, time.Now().Add(time.Minute).Unix())
			return
		}
		tokens++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	flow, err := client.BeginInstallationAuthorization(context.Background(), key, func(context.Context, InstallationProof) (string, error) {
		t.Error("expired intent reached signer")
		return "MAYCAQECAQE", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer flow.Close()
	flow.installation.proof.ExpiresAt = time.Now().Add(-time.Second).Unix()
	_, err = flow.Wait(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) || tokens != 0 {
		t.Fatal("expired intent reached exchange")
	}
	assertInstallationListenerClosed(t, flow.RedirectURI)
}

func TestInstallationKeyOrMissingSignerRefusesBeforeNetwork(t *testing.T) {
	network := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { network++ }))
	defer server.Close()
	client, _ := NewClient(configFor(server))
	for _, key := range []InstallationKey{{}, {Algorithm: "p256-sha256-x962-der-v1", PublicKey: "bad"}, func() InstallationKey { k := fixtureInstallationKey(); k.KeyFingerprint = "bad"; return k }()} {
		flow, err := client.BeginInstallationAuthorization(context.Background(), key, func(context.Context, InstallationProof) (string, error) { return "", nil })
		if flow != nil {
			flow.Close()
		}
		if err == nil || network != 0 {
			t.Fatal("unprotected key reached admission")
		}
	}
	flow, err := client.BeginInstallationAuthorization(context.Background(), fixtureInstallationKey(), nil)
	if flow != nil {
		flow.Close()
	}
	if err == nil || network != 0 {
		t.Fatal("missing signer reached admission")
	}
}

func installationCallback(t *testing.T, flow *Authorization, code string) {
	t.Helper()
	u, _ := url.Parse(flow.URL)
	r, err := http.Get(flow.RedirectURI + "?state=" + url.QueryEscape(u.Query().Get("state")) + "&code=" + url.QueryEscape(code))
	if err != nil {
		t.Fatal(err)
	}
	r.Body.Close()
	if r.StatusCode != http.StatusOK {
		t.Fatalf("callback status=%d", r.StatusCode)
	}
}

func assertInstallationListenerClosed(t *testing.T, address string) {
	t.Helper()
	u, err := url.Parse(address)
	if err != nil {
		t.Fatal(err)
	}
	connection, err := net.DialTimeout("tcp", u.Host, 100*time.Millisecond)
	if err == nil {
		connection.Close()
		t.Fatal("failed flow left callback listener open")
	}
}

func TestInstallationPolicyErrorsRemainDistinctFromReauthentication(t *testing.T) {
	for _, scenario := range []struct {
		status int
		body   string
	}{{403, `{"error":"installation_blocked","error_description":"private"}`}, {503, `{"error":"installation_policy_unknown","error_description":"private"}`}, {429, `{"error":"temporarily_limited","error_description":"private"}`}} {
		t.Run(fmt.Sprint(scenario.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(scenario.status)
				fmt.Fprint(w, scenario.body)
			}))
			defer server.Close()
			client, _ := NewClient(configFor(server))
			flow, err := client.BeginInstallationAuthorization(context.Background(), fixtureInstallationKey(), func(context.Context, InstallationProof) (string, error) { return "", nil })
			if flow != nil {
				flow.Close()
			}
			var typed *Error
			if !errors.As(err, &typed) || typed.StatusCode != scenario.status || errors.Is(err, ErrReauthRequired) {
				t.Fatal("installation policy lost its non-reauth status")
			}
		})
	}
}
