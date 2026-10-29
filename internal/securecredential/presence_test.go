package securecredential

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestConfirmRequiresExplicitEvidenceAndNoStorageFields(t *testing.T) {
	for _, body := range []string{
		`{"version":1,"ok":true}`, `{"version":1,"ok":true,"data":{}}`,
		`{"version":1,"ok":true,"data":{"present":true}}`,
		`{"version":1,"ok":true,"data":{"confirmed":false}}`,
		`{"version":1,"ok":true,"data":{"confirmed":true,"value":"fixture-value"}}`,
		`{"version":1,"ok":true,"data":{"confirmed":true,"value":null}}`,
		`{"version":1,"ok":true,"data":{"confirmed":true,"present":null}}`,
		`{"version":1,"ok":false,"error":{"code":"INVALID_REQUEST"}}`,
	} {
		s := testStore(func(context.Context, string, []byte) ([]byte, error) { return []byte(body), nil })
		if err := s.ConfirmUserPresence(context.Background()); !errors.Is(err, ErrUnavailable) {
			t.Fatalf("invalid confirmation accepted: %v", err)
		}
	}
}

func TestConfirmVerifiesSignerAndSendsNoCredential(t *testing.T) {
	verified, called := false, false
	s := testStore(func(_ context.Context, _ string, body []byte) ([]byte, error) {
		called = true
		var req request
		if json.Unmarshal(body, &req) != nil || !verified || req.Operation != "confirm" || !req.Interactive || req.Value != "" {
			t.Fatal("invalid confirmation request or ordering")
		}
		return []byte(`{"version":1,"ok":true,"data":{"confirmed":true}}`), nil
	})
	s.verify = func(context.Context, string) error { verified = true; return nil }
	if err := s.ConfirmUserPresence(context.Background()); err != nil || !called {
		t.Fatalf("confirmation failed: %v", err)
	}
	called = false
	s.verify = func(context.Context, string) error { return errors.New("private signature reason") }
	if err := s.ConfirmUserPresence(context.Background()); !errors.Is(err, ErrUnavailable) || called {
		t.Fatal("failed signature reached helper")
	}
}

func TestConfirmNoInputNeverInvokesVerifierOrHelper(t *testing.T) {
	s := testStore(func(context.Context, string, []byte) ([]byte, error) {
		t.Fatal("no-input invoked helper")
		return nil, nil
	})
	s.Interactive = false
	s.verify = func(context.Context, string) error { t.Fatal("no-input invoked verifier"); return nil }
	if err := s.ConfirmUserPresence(context.Background()); !errors.Is(err, ErrInteractionRequired) {
		t.Fatalf("no-input error=%v", err)
	}
}

func TestConfirmCancellationCannotBecomeSuccessfulEvidence(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := testStore(func(context.Context, string, []byte) ([]byte, error) {
		cancel()
		return []byte(`{"version":1,"ok":true,"data":{"confirmed":true}}`), nil
	})
	if err := s.ConfirmUserPresence(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("late success accepted: %v", err)
	}
	ctx, deadline := context.WithTimeout(context.Background(), time.Millisecond)
	defer deadline()
	s = testStore(func(ctx context.Context, _ string, _ []byte) ([]byte, error) {
		<-ctx.Done()
		return nil, ErrUnavailable
	})
	if err := s.ConfirmUserPresence(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline error=%v", err)
	}
}

func TestConfirmOwnDeadlineBoundsVerificationAndHelper(t *testing.T) {
	s := testStore(func(ctx context.Context, _ string, _ []byte) ([]byte, error) {
		if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > 2*time.Minute {
			t.Fatal("confirmation has no bounded deadline")
		}
		return []byte(`{"version":1,"ok":true,"data":{"confirmed":true}}`), nil
	})
	s.verify = func(ctx context.Context, _ string) error {
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("signature verification is outside confirmation deadline")
		}
		return nil
	}
	if err := s.ConfirmUserPresence(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestConfirmNativeTimeoutRemainsDeadline(t *testing.T) {
	s := testStore(func(context.Context, string, []byte) ([]byte, error) {
		return []byte(`{"version":1,"ok":false,"error":{"code":"TIMED_OUT","message":"private-timeout"}}`), nil
	})
	if err := s.ConfirmUserPresence(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("native timeout misclassified: %v", err)
	}
}
