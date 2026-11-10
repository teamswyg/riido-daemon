package standalonecli

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type transferTestKey struct{ *linkKeyFixture }

func (n *transferTestKey) SignDeviceTransferProof(_ context.Context, _ securecredential.InstallationPublicKey, p securecredential.DeviceTransferProof) (string, error) {
	n.signs++
	if p.PreviousLinkRevision != "7" || p.PreviousDeviceID != "source" {
		return "", securecredential.ErrUnavailable
	}
	return "MAYCAQECAQE", nil
}

func (n *transferTestKey) SignInstallationTransferRepairProof(_ context.Context, _ securecredential.InstallationPublicKey, p securecredential.InstallationTransferRepairProof) (string, error) {
	n.signs++
	if !linkRandom(p.OriginalRequestID) || p.RequestID == p.OriginalRequestID {
		return "", securecredential.ErrUnavailable
	}
	return "MAYCAQECAQE", nil
}

type transferTestDevice struct{ *linkDaemonFixture }

func (d transferTestDevice) ConfirmTransfer(ctx context.Context, p userauth.TransferProof, s string) error {
	if s != "MAYCAQECAQE" || p.ExpectedDeviceID != d.device.DeviceID {
		return securecredential.ErrUnavailable
	}
	d.confirms++
	return nil
}

func TestTransferFlowLostConfirmationResumesWithOriginalGrantAndReceipt(t *testing.T) {
	var f *linkFixture
	var sends atomic.Int32
	var repair atomic.Bool
	f = newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			t.Error("body missing")
		}
		j, err := f.store.loadTransfer(r.Context())
		if err != nil || j == nil {
			t.Error("network before custody")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		isRepair := strings.Contains(r.URL.Path, "/repairs")
		if isRepair != repair.Load() || r.Header.Get("Cookie") != "" || r.Header.Get("Origin") != "" {
			t.Error("wrong authorization path")
		}
		if isRepair {
			if r.Header.Get("Authorization") != "" || body["request_id"] != j.RepairRequest {
				t.Error("repair used incorrect custody")
			}
			if strings.HasSuffix(r.URL.Path, "/repairs") {
				if body["original_request_id"] != j.Request || body["recovery_secret"] != "" {
					t.Error("repair create lost original binding")
				}
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(map[string]any{"request_id": j.RepairRequest, "state": "pending", "expires_at": 1, "recovery_until": 2})
				return
			}
			linkState(w, j.RepairRequest, "complete")
			return
		}
		if r.Header.Get("Authorization") != "Bearer synthetic-access" || body["request_id"] != j.Request || body["recovery_secret"] != j.Secret {
			t.Error("wrong original grant/custody")
		}
		switch r.URL.Path {
		case "/oauth/cli/installation-transfers":
			sends.Add(1)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"request_id": j.Request, "state": "pending", "expires_at": 1})
		case "/oauth/cli/installation-transfers/execute":
			p := userauth.TransferProof{DeviceLinkProof: userauth.DeviceLinkProof{Resource: f.binding.Resource, ClientID: "riido-cli", InstallationID: "installation", IntentID: strings.Repeat("A", 43), ChallengeHash: strings.Repeat("b", 64), KeyFingerprint: f.binding.KeyFingerprint, ExpectedDeviceID: f.binding.ExpectedDeviceID, ApprovedPrincipalID: "human", ExpiresAt: time.Now().Add(time.Minute).UTC().Format(time.RFC3339Nano)}, PreviousDeviceID: "source", PreviousLinkRevision: "7"}
			_ = json.NewEncoder(w).Encode(userauth.TransferResult{RequestID: j.Request, State: "proof_required", Proof: &p})
		case "/oauth/cli/installation-transfers/recover":
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			t.Error("unexpected path")
			w.WriteHeader(http.StatusNotFound)
		}
	})
	n := &transferTestKey{f.native}
	n.key = securecredential.InstallationPublicKey{Algorithm: "p256-sha256-x962-der-v1", PublicKey: "BGsX0fLhLEJH-Lzm5WOkQPJ3A32BLeszoPShOUXYmMKWT-NC4v4af5uO5-tKfA-eFivOM1drMV7Oy7ZAaDe_UfU", KeyFingerprint: testRepairFingerprint()}
	f.binding.KeyFingerprint = n.key.KeyFingerprint
	f.store.Native = n
	f.store.TransferNative = &protectedNativeFixture{}
	f.store.RepairNative = &protectedNativeFixture{}
	d := transferTestDevice{f.daemon}
	if _, err := f.service.Transfer(t.Context(), d, true, false, false); err == nil || d.confirms != 1 || sends.Load() != 1 {
		t.Fatalf("ambiguous response handling: %v / %d", err, d.confirms)
	}
	before, err := f.store.loadTransfer(t.Context())
	if err != nil || before == nil || !before.Created {
		t.Fatal("lost original")
	}
	f.native.raw = nil
	repair.Store(true)
	loads := n.loads.Load()
	status, err := f.service.Transfer(t.Context(), nil, true, true, false)
	if err != nil || status.State != "complete" || n.loads.Load() != loads || n.ensures != 0 || sends.Load() != 1 {
		t.Fatalf("grantless repair failed: %v, %v", status, err)
	}
	after, err := f.store.loadTransfer(t.Context())
	if err != nil || after != nil {
		t.Fatal("completed journal retained")
	}
}

func TestTransferRepairRestartOnlyConsumesPositiveTerminalEvidence(t *testing.T) {
	var creates atomic.Int32
	f := newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/repairs") {
			creates.Add(1)
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	f.store.RepairNative = &protectedNativeFixture{}
	f.store.TransferNative = &protectedNativeFixture{}
	id, secret, _ := newTransferPair()
	rid, rsecret, _ := newTransferPair()
	j := transferJournal{Schema: "riido-cli-transfer.v1", Resource: f.binding.Resource, Origin: "browser", Session: f.binding.SessionID, Fingerprint: f.binding.KeyFingerprint, Device: f.binding.ExpectedDeviceID, Request: id, Secret: secret, Created: true, RepairRequest: rid, RepairSecret: rsecret}
	if err := f.store.saveTransfer(t.Context(), j); err != nil {
		t.Fatal(err)
	}
	if _, err := f.service.Transfer(t.Context(), nil, true, true, true); err == nil {
		t.Fatal("restart on ambiguous server acknowledgement")
	}
	saved, err := f.store.loadTransfer(t.Context())
	if err != nil || saved == nil || *saved != j || creates.Load() != 0 {
		t.Fatal("restart changed the original protected pair")
	}
}
