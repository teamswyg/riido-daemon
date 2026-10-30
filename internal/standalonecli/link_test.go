package standalonecli

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type linkDaemonFixture struct {
	device          LinkDevice
	infos, confirms int
	confirm         func() error
}

func (d *linkDaemonFixture) Info(context.Context) (LinkDevice, error) {
	d.infos++
	return d.device, nil
}

func (d *linkDaemonFixture) Confirm(_ context.Context, _ userauth.DeviceLinkProof, signature string) error {
	d.confirms++
	if signature != "synthetic-signature" {
		return errors.New("wrong signature")
	}
	if d.confirm != nil {
		return d.confirm()
	}
	return nil
}

type linkFixture struct {
	service    *Service
	httpClient *http.Client
	store      *ProtectedStore
	native     *linkKeyFixture
	binding    LinkBinding
	daemon     *linkDaemonFixture
}

func newLinkFixture(t *testing.T, handler http.HandlerFunc) *linkFixture {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	store, native, _, binding := newProtectedFixture(t)
	binding.Resource = server.URL
	store.Resource = server.URL
	keyNative := &linkKeyFixture{protectedNativeFixture: native, key: securecredential.InstallationPublicKey{KeyFingerprint: binding.KeyFingerprint}}
	store.Native = keyNative
	service, err := New(Config{APIOrigin: server.URL, Store: store, HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	return &linkFixture{service: service, httpClient: server.Client(), store: store, native: keyNative, binding: binding, daemon: &linkDaemonFixture{device: LinkDevice{Resource: server.URL, DeviceID: binding.ExpectedDeviceID}}}
}

func (f *linkFixture) prepare(t *testing.T, phase LinkPhase) PendingLink {
	t.Helper()
	pending, err := f.store.BeginPendingLink(t.Context(), f.binding)
	if err == nil && phase == LinkCreated {
		pending, err = f.store.MarkPendingLinkCreated(t.Context(), pending)
	}
	if err != nil {
		t.Fatal(err)
	}
	return pending
}

func linkRequest(t *testing.T, r *http.Request) map[string]string {
	t.Helper()
	var body map[string]string
	if r.Method != http.MethodPost || json.NewDecoder(r.Body).Decode(&body) != nil {
		t.Error("invalid fixed POST")
	}
	if r.URL.RawQuery != "" || r.Header.Get("Cookie") != "" || r.Header.Get("Origin") != "" {
		t.Error("unexpected ambient credentials")
	}
	if strings.HasSuffix(r.URL.Path, "/recover") && r.Header.Get("Authorization") != "" {
		t.Error("recovery used OAuth")
	}
	return body
}

func linkState(w http.ResponseWriter, id, state string) {
	_ = json.NewEncoder(w).Encode(map[string]string{"request_id": id, "state": state})
}

func linkProof(w http.ResponseWriter, id string, b LinkBinding) {
	p := userauth.DeviceLinkProof{Resource: b.Resource, ClientID: b.ClientID, InstallationID: "installation-original", IntentID: strings.Repeat("A", 43), ChallengeHash: strings.Repeat("b", 64), KeyFingerprint: b.KeyFingerprint, ExpectedDeviceID: b.ExpectedDeviceID, ApprovedPrincipalID: "owner-original", ExpiresAt: time.Now().Add(time.Minute).UTC().Format(time.RFC3339Nano)}
	_ = json.NewEncoder(w).Encode(map[string]any{"request_id": id, "state": "proof_required", "proof": p})
}

func assertPending(t *testing.T, f *linkFixture, original PendingLink) {
	t.Helper()
	pending, err := f.store.LoadPendingLink(t.Context(), f.binding.SessionID)
	if err != nil || pending == nil || !pending.sameRequest(original) {
		t.Fatal("original pending not preserved", err)
	}
	if f.native.ensures != 0 || f.native.deletes != 0 {
		t.Fatal("link mutated installation identity or OAuth")
	}
}

func TestLinkEndToEndWaitsForOriginalReceipt(t *testing.T) {
	var state atomic.Int32
	var creates atomic.Int32
	var f *linkFixture
	f = newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
		body := linkRequest(t, r)
		switch r.URL.Path {
		case "/oauth/cli/installation-links":
			creates.Add(1)
			pending, err := f.store.LoadPendingLink(r.Context(), f.binding.SessionID)
			if err != nil || pending == nil || pending.RequestID != body["request_id"] || pending.RecoverySecret != body["recovery_secret"] || r.Header.Get("Authorization") != "Bearer synthetic-access" {
				t.Error("create did not use persisted custody")
			}
			state.Store(1)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"request_id": body["request_id"], "state": "pending", "expires_at": time.Now().Add(time.Minute).Unix()})
		case "/oauth/cli/installation-links/recover":
			switch state.Load() {
			case 0:
				w.WriteHeader(http.StatusBadRequest)
			case 1:
				linkState(w, body["request_id"], "awaiting_approval")
			case 2:
				linkProof(w, body["request_id"], f.binding)
			case 3:
				linkState(w, body["request_id"], "linked")
			}
		default:
			t.Error("unexpected API")
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	f.service.openBrowser = func(ctx context.Context, address string) error {
		pending, err := f.store.LoadPendingLink(ctx, f.binding.SessionID)
		if err != nil || pending == nil || address != f.binding.Resource+"/oauth/cli/installation-links/"+pending.RequestID+"/authorize" || strings.Contains(address, pending.RecoverySecret) {
			t.Fatal("unsafe approval entry")
		}
		state.Store(2)
		return nil
	}
	f.daemon.confirm = func() error { state.Store(3); return nil }
	ctx, cancel := context.WithTimeout(t.Context(), 8*time.Second)
	defer cancel()
	result, err := f.service.Link(ctx, f.daemon, true)
	if err != nil || !result.Linked || creates.Load() != 1 || f.daemon.confirms != 1 || f.native.signs != 1 || f.native.ensures != 0 {
		t.Fatal("link did not complete original request", err)
	}
	pending, err := f.store.LoadPendingLink(t.Context(), f.binding.SessionID)
	if err != nil || pending != nil {
		t.Fatal("completed pending retained", err)
	}
}

func TestLinkReceiptRecoveryNeedsNoDaemonOrKey(t *testing.T) {
	f := newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
		body := linkRequest(t, r)
		linkState(w, body["request_id"], "linked")
	})
	f.prepare(t, LinkPrepared)
	f.native.keyErr = securecredential.ErrNotFound
	result, err := f.service.Link(t.Context(), nil, false)
	if err != nil || !result.Linked || f.native.existing != 0 || f.native.signs != 0 {
		t.Fatal("receipt recovery required live device/key", err)
	}
}

func TestLinkRecoveryFailuresNeverRecreateOrDelete(t *testing.T) {
	for _, test := range []struct {
		name   string
		phase  LinkPhase
		status int
	}{
		{"created400", LinkCreated, 400}, {"prepared503", LinkPrepared, 503}, {"created503", LinkCreated, 503}, {"prepared401", LinkPrepared, 401}, {"prepared409", LinkPrepared, 409}, {"prepared410", LinkPrepared, 410},
	} {
		t.Run(test.name, func(t *testing.T) {
			var calls atomic.Int32
			f := newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				linkRequest(t, r)
				if !strings.HasSuffix(r.URL.Path, "/recover") {
					t.Error("unexpected recreate")
				}
				w.WriteHeader(test.status)
			})
			original := f.prepare(t, test.phase)
			if _, err := f.service.Link(t.Context(), f.daemon, true); err == nil {
				t.Fatal("failed recovery accepted")
			}
			assertPending(t, f, original)
			if calls.Load() != 1 || f.daemon.infos != 0 || f.native.existing != 0 {
				t.Fatal("recovery depended on device/key")
			}
		})
	}
}
