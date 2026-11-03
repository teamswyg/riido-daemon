package standalonecli

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
)

type repairOperationFailure struct {
	*protectedNativeFixture
	err error
}

func (f repairOperationFailure) TryLock(context.Context) (func(), error) { return nil, f.err }

func TestRepairOperationRejectsUnavailableOrMissingUnlockBeforeEffects(t *testing.T) {
	for _, failure := range []error{nil, securecredential.ErrUnavailable, context.Canceled} {
		f, n, custody := newRepairFixture(t, func(http.ResponseWriter, *http.Request) {
			t.Error("unavailable lock called API")
		})
		f.store.RepairNative = repairOperationFailure{custody, failure}
		_, err := f.service.Repair(t.Context(), f.daemon, true, false)
		want := failure
		if want == nil {
			want = securecredential.ErrUnavailable
		}
		if !errors.Is(err, want) || custody.saves != 0 || n.existing != 0 || n.signs != 0 || f.daemon.confirms != 0 {
			t.Fatal("unavailable lock did not fail closed", err)
		}
	}
}

func TestRepairOperationReleasesAfterFailureAndCancellation(t *testing.T) {
	calls := 0
	f, _, custody := newRepairFixture(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		repairServerError(w, 503, "unavailable")
	})
	p, err := f.store.BeginPendingRepair(t.Context(), testRepairFingerprint())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := f.service.Repair(ctx, nil, true, false); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled repair lost cancellation", err)
	}
	for range 2 {
		if _, err := f.service.Repair(t.Context(), nil, true, false); err == nil {
			t.Fatal("server failure reported success")
		}
		unlock, err := custody.TryLock(t.Context())
		if err != nil {
			t.Fatal("failure retained operation lock", err)
		}
		unlock()
	}
	after, err := f.store.LoadPendingRepair(t.Context())
	if err != nil || after == nil || !after.sameRequest(p) || calls != 2 {
		t.Fatal("retry replaced request or was still busy", err)
	}
}
