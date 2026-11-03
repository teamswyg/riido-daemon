package standalonecli

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestRepairTerminalWithoutRestartPreservesOriginalCustody(t *testing.T) {
	for _, test := range []struct{ state, code string }{
		{"consent_rejected", "REPAIR_CONSENT_REJECTED"},
		{"recovery_expired", "REPAIR_RECOVERY_EXPIRED"},
	} {
		t.Run(test.state, func(t *testing.T) {
			for _, interactive := range []bool{false, true} {
				calls := 0
				f, native, _ := newRepairFixture(t, func(w http.ResponseWriter, r *http.Request) {
					calls++
					if r.URL.Path != "/oauth/cli/installation-repairs/recover" {
						t.Fatal("terminal state created a new request")
					}
					body := linkRequest(t, r)
					linkState(w, body["request_id"], test.state)
				})
				original, err := f.store.BeginPendingRepair(t.Context(), testRepairFingerprint())
				if err == nil {
					_, err = f.store.MarkPendingRepairCreated(t.Context(), original)
				}
				if err != nil {
					t.Fatal(err)
				}
				f.service.openBrowser = func(context.Context, string) error { t.Fatal("terminal state opened consent"); return nil }
				_, err = f.service.Repair(t.Context(), f.daemon, interactive, false)
				var failure *userauth.InstallationRepairError
				if !errors.As(err, &failure) || failure.Code != test.code {
					t.Fatal("terminal reason lost", err)
				}
				after, err := f.store.LoadPendingRepair(t.Context())
				if err != nil || after == nil || !after.sameRequest(original) || calls != 1 || native.repairSigns != 0 || native.signs != 0 || f.daemon.confirms != 0 {
					t.Fatal("terminal reply changed original custody or authority", err)
				}
			}
		})
	}
}
