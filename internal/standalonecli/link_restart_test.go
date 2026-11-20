package standalonecli

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestLinkExpiryKeepsOriginalWithoutExplicitRestart(t *testing.T) {
	f := newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
		body := linkRequest(t, r)
		if !strings.HasSuffix(r.URL.Path, "/recover") {
			t.Error("unexpected create")
		}
		linkState(w, body["request_id"], "approval_expired")
	})
	p := f.prepare(t, LinkCreated)
	_, err := f.service.Link(t.Context(), f.daemon, true)
	var failure *userauth.DeviceLinkError
	if !errors.As(err, &failure) || failure.Code != "LINK_APPROVAL_EXPIRED" {
		t.Fatal("expiry lacks actionable error", err)
	}
	assertPending(t, f, p)
}

func TestLinkRestartOnlyReplacesPositiveTerminalCustody(t *testing.T) {
	for _, state := range []string{"approval_expired", "awaiting_approval", "proof_expired", "linked", "unavailable"} {
		t.Run(state, func(t *testing.T) {
			var original string
			creates := 0
			f := newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) {
				body := linkRequest(t, r)
				if !strings.HasSuffix(r.URL.Path, "/recover") {
					creates++
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				if body["request_id"] != original {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if state == "unavailable" {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				linkState(w, original, state)
			})
			p := f.prepare(t, LinkCreated)
			original = p.RequestID
			before, err := f.store.Load(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			_, err = f.service.LinkWithRestart(t.Context(), f.daemon, true, true)
			after, loadErr := f.store.LoadPendingLink(t.Context(), f.binding.SessionID)
			if loadErr != nil {
				t.Fatal(loadErr)
			}
			credential, loadErr := f.store.Load(t.Context())
			if loadErr != nil || credential.SessionID != before.SessionID || credential.AccessToken != before.AccessToken || credential.RefreshToken != before.RefreshToken {
				t.Fatal("restart changed login")
			}
			switch state {
			case "approval_expired":
				if err == nil || after == nil || after.sameRequest(p) || after.LinkBinding != p.LinkBinding || creates != 1 {
					t.Fatal("terminal request not durably replaced", err)
				}
			case "linked":
				if err != nil || after != nil || creates != 0 {
					t.Fatal("linked receipt restarted", err)
				}
			default:
				if err == nil || after == nil || !after.sameRequest(p) || creates != 0 {
					t.Fatal("nonterminal request replaced", err)
				}
			}
		})
	}
}

func TestLinkRestartAtomicSaveAndStaleRequest(t *testing.T) {
	for _, committed := range []bool{false, true} {
		t.Run(map[bool]string{false: "before-commit", true: "after-commit"}[committed], func(t *testing.T) {
			f := newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected network") })
			p := f.prepare(t, LinkCreated)
			f.native.saveErr, f.native.commitOnFail = securecredential.ErrUnavailable, committed
			if _, err := f.store.RestartPendingLink(t.Context(), p); err == nil {
				t.Fatal("save failure hidden")
			}
			f.native.saveErr = nil
			after, err := f.store.LoadPendingLink(t.Context(), p.SessionID)
			if err != nil || after == nil || after.LinkBinding != p.LinkBinding || after.sameRequest(p) == committed {
				t.Fatal("partial or lost custody", err)
			}
			if committed {
				if _, err := f.store.RestartPendingLink(t.Context(), p); err == nil {
					t.Fatal("stale request replaced newer custody")
				}
			}
		})
	}
}

func TestLinkRestartNoninteractiveDoesNotReadCredentials(t *testing.T) {
	f := newLinkFixture(t, func(w http.ResponseWriter, r *http.Request) { t.Error("unexpected network") })
	p := f.prepare(t, LinkCreated)
	if _, err := f.service.LinkWithRestart(t.Context(), f.daemon, false, true); err == nil {
		t.Fatal("noninteractive restart accepted")
	}
	assertPending(t, f, p)
	if f.daemon.infos != 0 {
		t.Fatal("noninteractive restart contacted daemon")
	}
}
