package standalonecli

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestLinkLostCreateResponseRecoversOriginalPairAfterRestart(t *testing.T) {
	var created atomic.Bool
	var creates atomic.Int32
	f := newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
		body := linkRequest(t, r)
		if !strings.HasSuffix(r.URL.Path, "/recover") {
			creates.Add(1)
			created.Store(true)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		if !created.Load() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		linkState(w, body["request_id"], "linked")
	})
	original := f.prepare(t, LinkPrepared)
	if _, err := f.service.Link(t.Context(), f.daemon, true); err == nil {
		t.Fatal("lost create ACK reported success")
	}
	assertPending(t, f, original)
	restarted, err := New(Config{APIOrigin: f.binding.Resource, Store: f.store, HTTPClient: f.httpClient})
	if err != nil {
		t.Fatal(err)
	}
	result, err := restarted.Link(t.Context(), nil, false)
	if err != nil || !result.Linked || creates.Load() != 1 || f.native.existing != 0 {
		t.Fatal("restart regenerated request or needed device", err)
	}
}

func TestLinkConfirmAcknowledgementIsNotCompletion(t *testing.T) {
	var f *linkFixture
	f = newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
		body := linkRequest(t, r)
		linkProof(w, body["request_id"], f.binding)
	})
	original := f.prepare(t, LinkCreated)
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	result, err := f.service.Link(ctx, f.daemon, true)
	if !errors.Is(err, context.DeadlineExceeded) || result.Linked || f.daemon.confirms != 1 {
		t.Fatal("daemon ACK became final receipt", err)
	}
	assertPending(t, f, original)
}

func TestLinkLostConfirmAcknowledgementUsesOriginalReceipt(t *testing.T) {
	var linked atomic.Bool
	var f *linkFixture
	f = newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
		body := linkRequest(t, r)
		if linked.Load() {
			linkState(w, body["request_id"], "linked")
		} else {
			linkProof(w, body["request_id"], f.binding)
		}
	})
	f.prepare(t, LinkCreated)
	f.daemon.confirm = func() error { linked.Store(true); return errors.New("synthetic lost ACK") }
	result, err := f.service.Link(t.Context(), f.daemon, true)
	if err != nil || !result.Linked || f.daemon.confirms != 1 {
		t.Fatal("lost confirm ACK did not recover", err)
	}
}

func TestLinkNoninteractiveAndExpiredRemainPending(t *testing.T) {
	for _, state := range []string{"awaiting_approval", "proof_required", "proof_expired"} {
		t.Run(state, func(t *testing.T) {
			var f *linkFixture
			f = newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
				body := linkRequest(t, r)
				if state == "proof_required" {
					linkProof(w, body["request_id"], f.binding)
				} else {
					linkState(w, body["request_id"], state)
				}
			})
			original := f.prepare(t, LinkCreated)
			f.service.openBrowser = func(context.Context, string) error { t.Error("noninteractive opened browser"); return nil }
			if _, err := f.service.Link(t.Context(), f.daemon, false); err == nil {
				t.Fatal("unapproved request completed")
			}
			assertPending(t, f, original)
			if f.daemon.infos != 0 || f.native.existing != 0 || f.native.signs != 0 {
				t.Fatal("noninteractive requested live proof")
			}
		})
	}
}

func TestLinkChangedDeviceOrKeyCannotConfirm(t *testing.T) {
	for _, kind := range []string{"device", "key"} {
		t.Run(kind, func(t *testing.T) {
			var f *linkFixture
			f = newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
				body := linkRequest(t, r)
				linkProof(w, body["request_id"], f.binding)
			})
			original := f.prepare(t, LinkCreated)
			if kind == "device" {
				f.daemon.device.DeviceID = "replacement-device"
			} else {
				f.native.key.KeyFingerprint = strings.Repeat("c", 64)
			}
			if _, err := f.service.Link(t.Context(), f.daemon, true); err == nil {
				t.Fatal("changed binding completed")
			}
			assertPending(t, f, original)
			if f.native.signs != 0 || f.daemon.confirms != 0 {
				t.Fatal("changed binding signed or confirmed")
			}
		})
	}
}

func TestLinkLateReceiptCannotClearReplacementSession(t *testing.T) {
	var f *linkFixture
	f = newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
		body := linkRequest(t, r)
		credential := userauth.Credential{SessionID: "replacement-session", AccessToken: "synthetic-new", RefreshToken: "synthetic-new-refresh", TokenType: "Bearer", Scope: "read", ExpiresAt: time.Now().Add(time.Hour)}
		if err := saveProtectedCredential(t, f.store, credential); err != nil {
			t.Error(err)
		}
		linkState(w, body["request_id"], "linked")
	})
	f.prepare(t, LinkCreated)
	if _, err := f.service.Link(t.Context(), nil, false); !errors.Is(err, userauth.ErrLoggedOut) {
		t.Fatal("late receipt accepted after replacement", err)
	}
	credential, err := f.store.Load(t.Context())
	if err != nil || credential.SessionID != "replacement-session" || f.native.deletes != 0 {
		t.Fatal("late receipt modified replacement", err)
	}
}

func TestLinkPreparedNoInputNeverCreates(t *testing.T) {
	var requests atomic.Int32
	f := newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		linkRequest(t, r)
		w.WriteHeader(http.StatusBadRequest)
	})
	original := f.prepare(t, LinkPrepared)
	if _, err := f.service.Link(t.Context(), nil, false); err == nil {
		t.Fatal("noninteractive created")
	}
	if requests.Load() != 1 {
		t.Fatal("request recreated")
	}
	assertPending(t, f, original)
}

func TestLinkInvalidProofNeverReachesSigner(t *testing.T) {
	f := newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
		body := linkRequest(t, r)
		_ = json.NewEncoder(w).Encode(map[string]any{"request_id": body["request_id"], "state": "proof_required", "proof": map[string]string{"resource": "https://foreign.example.test"}})
	})
	original := f.prepare(t, LinkCreated)
	if _, err := f.service.Link(t.Context(), f.daemon, true); err == nil {
		t.Fatal("invalid proof accepted")
	}
	assertPending(t, f, original)
	if f.native.signs != 0 || f.daemon.infos != 0 {
		t.Fatal("invalid proof reached device")
	}
}

func TestLinkApprovalPollingDoesNotRepeatOSCredentialReads(t *testing.T) {
	var calls atomic.Int32
	var observedLoads atomic.Int32
	var f *linkFixture
	f = newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
		body := linkRequest(t, r)
		count := calls.Add(1)
		if count == 2 {
			observedLoads.Store(f.native.loads.Load())
		}
		if count == 3 {
			if got := f.native.loads.Load(); got != observedLoads.Load() {
				t.Errorf("unchanged approval poll read OS credential again: %d -> %d", observedLoads.Load(), got)
			}
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		linkState(w, body["request_id"], "awaiting_approval")
	})
	original := f.prepare(t, LinkCreated)
	f.service.openBrowser = func(context.Context, string) error { return nil }
	if _, err := f.service.Link(t.Context(), f.daemon, true); err == nil {
		t.Fatal("unapproved request succeeded")
	}
	assertPending(t, f, original)
}
