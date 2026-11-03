package standalonecli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
)

type repairKeyFixture struct {
	*linkKeyFixture
	repairSigns int
	repairProof securecredential.InstallationRepairProof
}

func (n *repairKeyFixture) SignInstallationRepairProof(_ context.Context, _ securecredential.InstallationPublicKey, p securecredential.InstallationRepairProof) (string, error) {
	n.repairSigns++
	n.repairProof = p
	return "MAYCAQECAQE", nil
}

func newRepairFixture(t *testing.T, handler http.HandlerFunc) (*linkFixture, *repairKeyFixture, *protectedNativeFixture) {
	t.Helper()
	f := newLinkFixture(t, handler)
	f.native.raw = nil
	f.native.key = securecredential.InstallationPublicKey{Algorithm: "p256-sha256-x962-der-v1", PublicKey: "BGsX0fLhLEJH-Lzm5WOkQPJ3A32BLeszoPShOUXYmMKWT-NC4v4af5uO5-tKfA-eFivOM1drMV7Oy7ZAaDe_UfU", KeyFingerprint: testRepairFingerprint()}
	f.binding.KeyFingerprint = testRepairFingerprint()
	n := &repairKeyFixture{linkKeyFixture: f.native}
	f.store.Native = n
	r := &protectedNativeFixture{}
	f.store.RepairNative = r
	return f, n, r
}

func repairServerError(w http.ResponseWriter, status int, code string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code, "error_description": "not logged"})
}

func TestRepairFullFlowPersistsBeforeCreateAndWaitsForReceipt(t *testing.T) {
	var f *linkFixture
	var n *repairKeyFixture
	var state, creates atomic.Int32
	f, n, _ = newRepairFixture(t, func(w http.ResponseWriter, r *http.Request) {
		body := linkRequest(t, r)
		if r.Header.Get("Authorization") != "" {
			t.Error("repair loaded OAuth")
		}
		if strings.HasSuffix(r.URL.Path, "/recover") {
			switch state.Load() {
			case 0:
				repairServerError(w, 400, "invalid_installation_repair")
			case 1:
				linkState(w, body["request_id"], "awaiting_approval")
			case 2:
				linkProof(w, body["request_id"], f.binding)
			case 3:
				linkState(w, body["request_id"], "linked")
			}
			return
		}
		if r.URL.Path != "/oauth/cli/installation-repairs" {
			t.Error("unexpected API")
		}
		p, err := f.store.LoadPendingRepair(r.Context())
		if err != nil || p == nil {
			t.Error("not durable before create")
			return
		}
		digest := sha256.Sum256([]byte(p.RecoverySecret))
		if body["request_id"] != p.RequestID || body["recovery_hash"] != hex.EncodeToString(digest[:]) || body["recovery_secret"] != "" {
			t.Error("wrong request custody")
		}
		creates.Add(1)
		state.Store(1)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"request_id": p.RequestID, "state": "pending", "expires_at": 1, "recovery_until": 2})
	})
	f.service.openBrowser = func(ctx context.Context, address string) error {
		p, err := f.store.LoadPendingRepair(ctx)
		if err != nil || p == nil || address != f.binding.Resource+"/oauth/cli/installation-repairs/"+p.RequestID+"/authorize" || strings.Contains(address, p.RecoverySecret) {
			t.Error("wrong browser entry")
		}
		state.Store(2)
		return nil
	}
	f.daemon.confirm = func() error { state.Store(3); return errors.New("lost ACK") }
	ctx, cancel := context.WithTimeout(t.Context(), 8*time.Second)
	defer cancel()
	result, err := f.service.Repair(ctx, f.daemon, true, false)
	if err != nil || !result.Linked || creates.Load() != 1 || n.repairSigns != 1 || f.daemon.confirms != 1 || n.ensures != 0 || n.saves != 0 || n.deletes != 0 {
		t.Fatal("repair did not finish without OAuth", err)
	}
	if p, err := f.store.LoadPendingRepair(ctx); err != nil || p != nil {
		t.Fatal("receipt did not clear original request", err)
	}
}

func TestRepairReceiptRecoveryNeedsNoDaemonKeyOrBrowser(t *testing.T) {
	f, n, _ := newRepairFixture(t, func(w http.ResponseWriter, r *http.Request) {
		body := linkRequest(t, r)
		linkState(w, body["request_id"], "linked")
	})
	_, err := f.store.BeginPendingRepair(t.Context(), testRepairFingerprint())
	if err != nil {
		t.Fatal(err)
	}
	n.keyErr = securecredential.ErrNotFound
	result, err := f.service.Repair(t.Context(), nil, false, false)
	if err != nil || !result.Linked || n.existing != 0 || n.repairSigns != 0 || n.signs != 0 {
		t.Fatal("receipt required active daemon", err)
	}
}

func TestRepairUnknownRecoveryNeverRegeneratesCreatedRequest(t *testing.T) {
	for _, test := range []struct {
		status int
		code   string
	}{{400, "invalid_request"}, {400, "invalid_installation_repair"}, {409, "installation_policy_unknown"}, {503, "unavailable"}} {
		t.Run(test.code, func(t *testing.T) {
			creates := 0
			f, n, _ := newRepairFixture(t, func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasSuffix(r.URL.Path, "/recover") {
					creates++
				}
				repairServerError(w, test.status, test.code)
			})
			p, err := f.store.BeginPendingRepair(t.Context(), testRepairFingerprint())
			if err == nil {
				_, err = f.store.MarkPendingRepairCreated(t.Context(), p)
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err = f.service.Repair(t.Context(), f.daemon, true, false); err == nil {
				t.Fatal("unknown reported success")
			}
			after, err := f.store.LoadPendingRepair(t.Context())
			if err != nil || after == nil || !after.sameRequest(p) || creates != 0 || n.repairSigns != 0 {
				t.Fatal("recreated ambiguous request", err)
			}
		})
	}
}

func TestRepairProofChecksAuthenticatedDeviceAndPreservesRequest(t *testing.T) {
	var f *linkFixture
	f, n, _ := newRepairFixture(t, func(w http.ResponseWriter, r *http.Request) {
		body := linkRequest(t, r)
		linkProof(w, body["request_id"], f.binding)
	})
	p, err := f.store.BeginPendingRepair(t.Context(), testRepairFingerprint())
	if err != nil {
		t.Fatal(err)
	}
	f.daemon.device.DeviceID = "other-device"
	if _, err = f.service.Repair(t.Context(), f.daemon, true, false); err == nil || n.signs != 0 || f.daemon.confirms != 0 {
		t.Fatal("foreign daemon reached signer")
	}
	after, err := f.store.LoadPendingRepair(t.Context())
	if err != nil || after == nil || !after.sameRequest(p) {
		t.Fatal("device mismatch lost original custody", err)
	}
}
