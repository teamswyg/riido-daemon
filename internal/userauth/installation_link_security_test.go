package userauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestDeviceLinkRecoveryValidatesEveryProofBinding(t *testing.T) {
	input := linkRecoveryInput(linkFixtureRequest())
	var proof DeviceLinkProof
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		linkReply(w, 200, DeviceLinkRecoveryResult{RequestID: input.RequestID, State: "proof_required", Proof: &proof})
	}))
	defer server.Close()
	client := linkFixtureClient(t, server)
	original := DeviceLinkProof{Resource: server.URL, ClientID: "riido-cli", InstallationID: "stored-installation", IntentID: base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("i", 32))), ChallengeHash: strings.Repeat("b", 64), KeyFingerprint: input.KeyFingerprint, ExpectedDeviceID: input.ExpectedDeviceID, ApprovedPrincipalID: "approved-human", ExpiresAt: time.Now().UTC().Add(time.Minute).Format(time.RFC3339Nano)}
	proof = original
	result, err := client.RecoverDeviceLink(context.Background(), input)
	if err != nil || result.Proof == nil || *result.Proof != original {
		t.Fatal("valid typed proof changed", err)
	}
	mutations := []func(*DeviceLinkProof){
		func(p *DeviceLinkProof) { p.Resource = "https://other.example.test" }, func(p *DeviceLinkProof) { p.ClientID = "other-client" },
		func(p *DeviceLinkProof) { p.InstallationID = "../other" }, func(p *DeviceLinkProof) { p.IntentID = "not-original-random" },
		func(p *DeviceLinkProof) { p.ChallengeHash = strings.Repeat("B", 64) }, func(p *DeviceLinkProof) { p.KeyFingerprint = strings.Repeat("c", 64) },
		func(p *DeviceLinkProof) { p.ExpectedDeviceID = "other-device" }, func(p *DeviceLinkProof) { p.ApprovedPrincipalID = "other/user" },
		func(p *DeviceLinkProof) { p.ExpiresAt = "2026-10-29T00:00:00+00:00" }, func(p *DeviceLinkProof) { p.ExpiresAt = time.Now().Add(-time.Minute).UTC().Format(time.RFC3339Nano) },
	}
	for _, mutate := range mutations {
		proof = original
		mutate(&proof)
		result, err = client.RecoverDeviceLink(context.Background(), input)
		if err == nil || result.Proof != nil {
			t.Fatal("invalid proof reached the signer boundary")
		}
	}
}

func TestDeviceLinkTransportRejectsRedirectWithoutLeakingSecret(t *testing.T) {
	targetCalls := 0
	target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { targetCalls++; w.WriteHeader(http.StatusOK) }))
	defer target.Close()
	origin := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/unexpected", http.StatusTemporaryRedirect)
	}))
	defer origin.Close()
	client := linkFixtureClient(t, origin)
	input := linkRecoveryInput(linkFixtureRequest())
	_, err := client.RecoverDeviceLink(context.Background(), input)
	var linkErr *DeviceLinkError
	if !errors.As(err, &linkErr) || linkErr.StatusCode != http.StatusTemporaryRedirect || targetCalls != 0 || strings.Contains(err.Error(), input.RecoverySecret) {
		t.Fatal("redirect was followed or exposed secret", err)
	}
	if client.http.Jar == nil {
		t.Fatal("link request mutated the ordinary OAuth client cookie policy")
	}
}

func TestDeviceLinkRetryAfterLostResponseReusesCallerOwnedPair(t *testing.T) {
	input := linkFixtureRequest()
	calls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var fields map[string]string
		if json.NewDecoder(r.Body).Decode(&fields) != nil || fields["request_id"] != input.RequestID || fields["recovery_secret"] != input.RecoverySecret {
			t.Error("response-loss retry replaced original pair")
		}
		if calls == 1 {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			_ = conn.Close()
			return
		}
		linkReply(w, 201, DeviceLinkApproval{RequestID: input.RequestID, State: "pending", ExpiresAt: 1})
	}))
	defer server.Close()
	store := linkFixtureStore()
	manager := NewManager(linkFixtureClient(t, server), store)
	if _, err := manager.CreateDeviceLink(context.Background(), "fixture-session", input); err == nil || calls != 1 {
		t.Fatal("lost response was silently retried")
	}
	if _, err := manager.CreateDeviceLink(context.Background(), "fixture-session", input); err != nil || calls != 2 {
		t.Fatal("explicit original retry failed", err)
	}
	if store.credential.SessionID != "fixture-session" || !store.present {
		t.Fatal("link retry revoked or replaced login")
	}
}

func TestDeviceLinkCreateRejectsUnboundInputAndGrantScope(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); t.Error("rejected input reached HTTP") }))
	defer server.Close()
	client := linkFixtureClient(t, server)
	store := linkFixtureStore()
	manager := NewManager(client, store)
	input := linkFixtureRequest()
	for _, session := range []string{"", "../bad"} {
		if _, err := manager.CreateDeviceLink(context.Background(), session, input); err == nil {
			t.Fatal("unbound session accepted")
		}
	}
	bad := input
	bad.RequestID = "invalid"
	if _, err := manager.CreateDeviceLink(context.Background(), "fixture-session", bad); err == nil {
		t.Fatal("invalid pair accepted")
	}
	if store.loads != 0 {
		t.Fatal("invalid request opened protected storage")
	}
	store.credential.Scope = "read write"
	if _, err := manager.CreateDeviceLink(context.Background(), "fixture-session", input); err == nil {
		t.Fatal("write-scoped credential accepted")
	}
	if calls.Load() != 0 {
		t.Fatal("unsafe request called server")
	}
}

func TestDeviceLinkRecoveryBoundsDeadlineAndPropagatesCancellation(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		close(entered)
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	defer server.Close()
	defer close(release)
	client := linkFixtureClient(t, server)
	client.http.Timeout = time.Hour
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := client.RecoverDeviceLink(ctx, linkRecoveryInput(linkFixtureRequest())); done <- err }()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal("cancellation hidden", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("cancellation did not bound request")
	}
}

func TestDeviceLinkRecoveryRejectsDuplicateNestedProof(t *testing.T) {
	input := linkRecoveryInput(linkFixtureRequest())
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"request_id":"`+input.RequestID+`","state":"proof_required","proof":{"resource":"https://bad.invalid","resource":"https://api.example.test"}}`)
	}))
	defer server.Close()
	_, err := linkFixtureClient(t, server).RecoverDeviceLink(context.Background(), input)
	var linkErr *DeviceLinkError
	if !errors.As(err, &linkErr) || linkErr.Code != "LINK_INVALID_RESPONSE" {
		t.Fatal("duplicate nested authority accepted", err)
	}
}
