package standalonecli

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestRepairRejectsConcurrentOperationBeforeEffects(t *testing.T) {
	var calls, approvals atomic.Int32
	var linked atomic.Bool
	f, n, _ := newRepairFixture(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		body := linkRequest(t, r)
		state := "awaiting_approval"
		if linked.Load() {
			state = "linked"
		}
		linkState(w, body["request_id"], state)
	})
	p, err := f.store.BeginPendingRepair(t.Context(), testRepairFingerprint())
	if err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	releaseHolder := func() { once.Do(func() { close(release) }) }
	defer releaseHolder()
	f.service.openBrowser = func(ctx context.Context, _ string) error {
		if approvals.Add(1) != 1 {
			return errors.New("duplicate approval")
		}
		close(entered)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-release:
			linked.Store(true)
			return nil
		}
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { _, e := f.service.Repair(ctx, f.daemon, true, false); done <- e }()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("holder did not reach approval")
	}
	_, err = f.service.Repair(ctx, f.daemon, true, false)
	var conflict *userauth.InstallationRepairError
	if !errors.As(err, &conflict) || conflict.Code != "REPAIR_IN_PROGRESS" || calls.Load() != 1 || approvals.Load() != 1 || n.signs != 0 || n.repairSigns != 0 || f.daemon.confirms != 0 {
		t.Error("competing repair performed work instead of rejecting", err)
	}
	after, loadErr := f.store.LoadPendingRepair(ctx)
	if loadErr != nil || after == nil || !after.sameRequest(p) {
		t.Error("competing repair changed custody", loadErr)
	}
	releaseHolder()
	if err := <-done; err != nil {
		t.Fatal("holder failed", err)
	}
}
