package standalonecli

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func repairStoreFixture(t *testing.T) (*ProtectedStore, *protectedNativeFixture, *protectedNativeFixture) {
	t.Helper()
	s, n, _, _ := newProtectedFixture(t)
	n.raw = nil
	r := &protectedNativeFixture{}
	s.RepairNative = r
	return s, n, r
}

func TestPendingRepairDurableWithoutCredentialAndAcrossRestart(t *testing.T) {
	s, n, r := repairStoreFixture(t)
	ctx := context.Background()
	fingerprint := testRepairFingerprint()
	r.saveErr = securecredential.ErrUnavailable
	r.commitOnFail = true
	if _, err := s.BeginPendingRepair(ctx, fingerprint); err == nil {
		t.Fatal("ambiguous native save reported success")
	}
	r.saveErr = nil
	p, err := (&ProtectedStore{Native: n, RepairNative: r, Resource: s.Resource}).BeginPendingRepair(ctx, fingerprint)
	if err != nil || p.SessionID != "" || p.Phase != LinkPrepared || !p.valid() {
		t.Fatal("original repair not recovered", err)
	}
	if r.saves != 1 {
		t.Fatal("retry regenerated request")
	}
	if _, err = s.Load(ctx); !errors.Is(err, userauth.ErrNotAuthenticated) {
		t.Fatal("repair became credential")
	}
	created, err := s.MarkPendingRepairCreated(ctx, p)
	if err != nil || created.Phase != LinkCreated {
		t.Fatal("phase failed", err)
	}
	if err = s.ClearPendingRepair(ctx, p); err != nil {
		t.Fatal("original identity cleanup rejected phase progress", err)
	}
	if len(r.raw) != 0 {
		t.Fatal("repair retained after exact clear")
	}
}

func TestPendingRepairSerializesAndRejectsReplacedSession(t *testing.T) {
	s, n, _ := repairStoreFixture(t)
	ctx := context.Background()
	results := make(chan PendingRepair, 8)
	errs := make(chan error, 8)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			p, e := s.BeginPendingRepair(ctx, testRepairFingerprint())
			results <- p
			errs <- e
		})
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var first PendingRepair
	for p := range results {
		if first.RequestID == "" {
			first = p
		}
		if !first.sameRequest(p) {
			t.Fatal("concurrency created multiple pairs")
		}
	}
	credential := userauth.Credential{SessionID: "other-session", AccessToken: "synthetic", TokenType: "Bearer"}
	raw, _ := encodeCredential(credential)
	n.mu.Lock()
	n.raw = raw
	n.mu.Unlock()
	if _, err := s.LoadPendingRepair(ctx); !errors.Is(err, userauth.ErrLoggedOut) {
		t.Fatal("account replacement kept old repair active", err)
	}
	if err := s.ClearPendingRepair(ctx, first); err == nil {
		t.Fatal("old request cleared after replacement")
	}
}

func TestPendingRepairLogoutCancelsEvenWithoutOAuth(t *testing.T) {
	s, _, r := repairStoreFixture(t)
	ctx := context.Background()
	p, err := s.BeginPendingRepair(ctx, testRepairFingerprint())
	if err != nil {
		t.Fatal(err)
	}
	service, err := New(Config{APIOrigin: s.Resource, Store: s})
	if err != nil {
		t.Fatal(err)
	}
	if err = service.Logout(ctx); err != nil {
		t.Fatal("credential-free logout failed", err)
	}
	if len(r.raw) != 0 {
		t.Fatal("logout left repair authority")
	}
	if err = s.ClearPendingRepair(ctx, p); err == nil {
		t.Fatal("late receipt succeeded after logout")
	}
}

func TestPendingRepairRejectsCorruptRecordWithoutOverwrite(t *testing.T) {
	s, _, r := repairStoreFixture(t)
	ctx := context.Background()
	p, err := s.BeginPendingRepair(ctx, testRepairFingerprint())
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if json.Unmarshal(r.raw, &body) != nil {
		t.Fatal("fixture")
	}
	body["schema_version"] = "future"
	bad, _ := json.Marshal(body)
	r.raw = bad
	if _, err = s.BeginPendingRepair(ctx, p.KeyFingerprint); err == nil {
		t.Fatal("unknown schema accepted")
	}
	if r.saves != 1 {
		t.Fatal("corrupt record overwritten")
	}
}

func testRepairFingerprint() string {
	return "698bea63dc44a344663ff1429aea10842df27b6b991ef25866b2c6c02cdcc5be"
}
