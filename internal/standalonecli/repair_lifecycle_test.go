package standalonecli

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
)

func TestRepairExplicitRestartRequiresTerminalState(t *testing.T) {
	for _, state := range []string{"approval_expired", "repair_expired", "consent_rejected", "recovery_expired", "awaiting_approval", "preparing", "linked"} {
		t.Run(state, func(t *testing.T) {
			var original string
			creates := 0
			f, n, _ := newRepairFixture(t, func(w http.ResponseWriter, r *http.Request) {
				body := linkRequest(t, r)
				if !strings.HasSuffix(r.URL.Path, "/recover") {
					creates++
					repairServerError(w, 503, "unavailable")
					return
				}
				if body["request_id"] != original {
					repairServerError(w, 400, "invalid_installation_repair")
					return
				}
				linkState(w, original, state)
			})
			p, err := f.store.BeginPendingRepair(t.Context(), testRepairFingerprint())
			if err != nil {
				t.Fatal(err)
			}
			original = p.RequestID
			_, err = f.service.Repair(t.Context(), f.daemon, true, true)
			after, loadErr := f.store.LoadPendingRepair(t.Context())
			if loadErr != nil {
				t.Fatal(loadErr)
			}
			switch state {
			case "approval_expired", "repair_expired", "consent_rejected", "recovery_expired":
				if err == nil || after == nil || after.sameRequest(p) || creates != 1 || n.repairSigns != 1 {
					t.Fatal("terminal restart lost new durable pair", err)
				}
			case "linked":
				if err != nil || after != nil || creates != 0 || n.repairSigns != 0 {
					t.Fatal("linked receipt restarted", err)
				}
			default:
				if err == nil || after == nil || !after.sameRequest(p) || creates != 0 || n.repairSigns != 0 {
					t.Fatal("active request replaced", err)
				}
			}
		})
	}
}

func TestRepairLogoutDuringApprovalPreservesCancellation(t *testing.T) {
	f, n, r := newRepairFixture(t, func(w http.ResponseWriter, request *http.Request) {
		body := linkRequest(t, request)
		linkState(w, body["request_id"], "awaiting_approval")
	})
	_, err := f.store.BeginPendingRepair(t.Context(), testRepairFingerprint())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	f.service.openBrowser = func(ctx context.Context, _ string) error {
		if err := f.service.Logout(ctx); err != nil {
			t.Fatal(err)
		}
		cancel()
		return nil
	}
	if _, err = f.service.Repair(ctx, f.daemon, true, false); err == nil {
		t.Fatal("logout allowed completion")
	}
	if len(r.raw) != 0 || n.signs != 0 || f.daemon.confirms != 0 {
		t.Fatal("logout resurrected request")
	}
}

func TestRepairNoninteractiveCannotApproveOrRestart(t *testing.T) {
	for _, restart := range []bool{false, true} {
		calls := 0
		f, n, _ := newRepairFixture(t, func(w http.ResponseWriter, r *http.Request) {
			calls++
			body := linkRequest(t, r)
			linkState(w, body["request_id"], "awaiting_approval")
		})
		p, err := f.store.BeginPendingRepair(t.Context(), testRepairFingerprint())
		if err != nil {
			t.Fatal(err)
		}
		f.service.openBrowser = func(context.Context, string) error { t.Fatal("noninteractive opened browser"); return nil }
		if _, err = f.service.Repair(t.Context(), f.daemon, false, restart); err == nil {
			t.Fatal("noninteractive mutation")
		}
		after, err := f.store.LoadPendingRepair(t.Context())
		if err != nil || after == nil || !after.sameRequest(p) || n.signs != 0 || n.repairSigns != 0 || restart && calls != 0 {
			t.Fatal("noninteractive changed custody")
		}
	}
}

type repairDeleteFailure struct{ *protectedNativeFixture }

func (*repairDeleteFailure) Delete(context.Context) error { return securecredential.ErrUnavailable }

type absentCredentialFixture struct{ *protectedNativeFixture }

func (*absentCredentialFixture) Status(context.Context) (bool, error) { return false, nil }

func TestRepairLocalLogoutFailureIsNotRemoteRevocationFailure(t *testing.T) {
	s, n, r := repairStoreFixture(t)
	s.Native = &absentCredentialFixture{n}
	s.RepairNative = &repairDeleteFailure{r}
	service, err := New(Config{APIOrigin: s.Resource, Store: s})
	if err != nil {
		t.Fatal(err)
	}
	err = service.Logout(t.Context())
	if err == nil || errors.Is(err, ErrLogoutRemoteIncomplete) {
		t.Fatal("local cleanup failure claimed local deletion", err)
	}
}

func TestRepairRestartPreservesCreatedRequestOnAmbiguousRejection(t *testing.T) {
	calls := 0
	f, n, _ := newRepairFixture(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		repairServerError(w, 400, "invalid_installation_repair")
	})
	p, err := f.store.BeginPendingRepair(t.Context(), testRepairFingerprint())
	if err == nil {
		_, err = f.store.MarkPendingRepairCreated(t.Context(), p)
	}
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.service.Repair(t.Context(), f.daemon, true, true); err == nil {
		t.Fatal("ambiguous rejection succeeded")
	}
	after, err := f.store.LoadPendingRepair(t.Context())
	if err != nil || after == nil || !after.sameRequest(p) || calls != 1 || n.repairSigns != 0 {
		t.Fatal("400 lost original recovery custody", err)
	}
}
