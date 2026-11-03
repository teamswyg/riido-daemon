package standalonecli

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/userapi"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestCompoundReadsHoldLockThroughResponseParsing(t *testing.T) {
	for _, kind := range []string{"context", "resources", "repositories"} {
		t.Run(kind, func(t *testing.T) {
			service, _, calls, _ := compoundFixture(t, kind, func(_ *Service, store *testStore, _ int) {
				if store.mu.TryLock() {
					store.mu.Unlock()
					t.Error("another process could replace login before parsing completed")
				}
			}, false)
			present, err := compoundRead(t.Context(), service, kind)
			if !present || err != nil || calls.Load() != 2 {
				t.Fatalf("same-session read failed: result=%v err=%v HTTP=%d", present, err, calls.Load())
			}
			if _, err := service.manager.AccessToken(t.Context()); err != nil {
				t.Fatal("successful read retained store lock", err)
			}
		})
	}
}

func TestCompoundReadsRejectMissingScopeBeforeCredentials(t *testing.T) {
	for _, ids := range [][2]string{{"", "team"}, {"work", ""}} {
		// A zero Service has no credential manager; invalid scope must return
		// before touching it, preserving the previous Reader validation order.
		service := &Service{}
		if got, err := service.Context(t.Context(), ids[0], ids[1]); got.RootID != "" || !errors.Is(err, userapi.ErrInvalidResponse) {
			t.Fatal("context accepted missing scope", err)
		}
		if got, err := service.Resources(t.Context(), ids[0], ids[1]); got != nil || !errors.Is(err, userapi.ErrInvalidResponse) {
			t.Fatal("resources accepted missing scope", err)
		}
		if got, err := service.Repositories(t.Context(), ids[0], ids[1]); got != nil || !errors.Is(err, userapi.ErrInvalidResponse) {
			t.Fatal("repositories accepted missing scope", err)
		}
	}
}

func TestCompoundReadsDiscardChangedLogin(t *testing.T) {
	for _, kind := range []string{"context", "resources", "repositories"} {
		for _, after := range []int{1, 2} {
			t.Run(fmt.Sprintf("%s/after-%d", kind, after), func(t *testing.T) {
				service, _, calls, _ := compoundFixture(t, kind, func(_ *Service, store *testStore, call int) {
					if call == after {
						store.credential.SessionID = "replacement-session"
						store.credential.AccessToken = "replacement-access"
					}
				}, false)
				present, err := compoundRead(t.Context(), service, kind)
				if present || !errors.Is(err, userauth.ErrLoggedOut) || calls.Load() != int32(after) {
					t.Fatalf("mixed login escaped: result=%v err=%v HTTP=%d", present, err, calls.Load())
				}
				if _, err := service.manager.AccessToken(t.Context()); err != nil {
					t.Fatal("read did not release store lock", err)
				}
			})
		}
	}
}

func TestCompoundReadsKeepLoginThroughRefresh(t *testing.T) {
	for _, kind := range []string{"context", "resources", "repositories"} {
		t.Run(kind, func(t *testing.T) {
			service, store, calls, refreshes := compoundFixture(t, kind, nil, true)
			present, err := compoundRead(t.Context(), service, kind)
			if !present || err != nil || calls.Load() != 3 || refreshes.Load() != 1 || store.credential.SessionID != "fixture-session" {
				t.Fatalf("original login/refresh failed: result=%v err=%v HTTP=%d refresh=%d", present, err, calls.Load(), refreshes.Load())
			}
		})
	}
}

func TestCompoundReadsCancelAfterParsing(t *testing.T) {
	for _, kind := range []string{"context", "resources", "repositories"} {
		for _, after := range []int{1, 2} {
			t.Run(fmt.Sprintf("%s/after-%d", kind, after), func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				service, _, calls, _ := compoundFixture(t, kind, func(_ *Service, _ *testStore, call int) {
					if call == after {
						cancel()
					}
				}, false)
				present, err := compoundRead(ctx, service, kind)
				if present || !errors.Is(err, context.Canceled) || calls.Load() != int32(after) {
					t.Fatalf("cancelled read emitted data: result=%v err=%v HTTP=%d", present, err, calls.Load())
				}
				if _, err := service.manager.AccessToken(t.Context()); err != nil {
					t.Fatal("cancelled read retained lock", err)
				}
			})
		}
	}
}
