package securecredential

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

type failingHelperSession struct {
	err    error
	closed bool
}

func (s *failingHelperSession) roundTrip(context.Context, []byte) ([]byte, error) {
	return []byte("synthetic-sensitive-output"), s.err
}

func (s *failingHelperSession) close() { s.closed = true }

func TestHelperInvocationFailurePreservesOnlyContextCause(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want error
	}{
		{"deadline", fmt.Errorf("synthetic-sensitive-diagnostic: %w", context.DeadlineExceeded), context.DeadlineExceeded},
		{"cancel", fmt.Errorf("synthetic-sensitive-diagnostic: %w", context.Canceled), context.Canceled},
		{"native-failure", errors.New("synthetic-sensitive-diagnostic"), ErrUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, shared := range []bool{false, true} {
				started := withFakeSessions(t)
				broken := &failingHelperSession{err: tc.err}
				if shared {
					sessions["/fixture/helper"] = broken
				}
				store := sessionStore(shared, true)
				if !shared {
					store.run = func(context.Context, string, []byte) ([]byte, error) {
						return []byte("synthetic-sensitive-output"), tc.err
					}
				}
				value, err := store.Load(context.Background())
				if !errors.Is(err, tc.want) || err.Error() != tc.want.Error() || len(value) != 0 {
					t.Fatalf("shared=%v: failure did not preserve only the public cause", shared)
				}
				if shared {
					if !broken.closed {
						t.Fatal("failed helper remained reusable")
					}
					value, err = store.Load(context.Background())
					if err != nil || string(value) != "fixture" || len(*started) != 1 {
						t.Fatal("retry did not start a fresh verified helper session")
					}
				}
				CloseSessions()
			}
		})
	}
}
