package securecredential

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func portableHelperFixture(t *testing.T, mode string) (string, string) {
	t.Helper()
	path, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	capture := filepath.Join(t.TempDir(), "captured")
	t.Setenv(processFixtureMode, mode)
	t.Setenv(processFixtureCapture, capture)
	return path, capture
}

func waitForEmptyCapture(ctx context.Context, capture string) error {
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	var observedAt time.Time
	for {
		data, err := os.ReadFile(capture)
		if err == nil {
			if len(data) != 0 {
				return errors.New("request bytes preceded process approval")
			}
			if observedAt.IsZero() {
				observedAt = time.Now()
			}
			// Give the independent pipe peer a bounded chance to expose any
			// bytes delivered before approval. Mutation checks cover this order.
			if time.Since(observedAt) >= 100*time.Millisecond {
				return nil
			}
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func assertNoCapturedRequest(t *testing.T, capture string) {
	t.Helper()
	data, err := os.ReadFile(capture)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Fatal("unverified child received request bytes")
	}
}

func TestPortableHelperDeliversAfterProcessApproval(t *testing.T) {
	path, capture := portableHelperFixture(t, "echo")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	approved := false
	input := []byte("synthetic-request")
	output, err := runVerifiedHelper(ctx, path, input, func(ctx context.Context, pid int) error {
		if pid <= 0 {
			return errors.New("missing child identity")
		}
		if err := waitForEmptyCapture(ctx, capture); err != nil {
			return err
		}
		approved = true
		return nil
	})
	if err != nil || !approved || string(output) != "fixture-result" {
		t.Fatalf("approved helper failed: %v", err)
	}
	data, err := os.ReadFile(capture)
	if err != nil || !bytes.Equal(data, input) {
		t.Fatal("approved helper did not receive the exact request")
	}
}

func TestPortableHelperDenialAndCancellationSendNoInput(t *testing.T) {
	for _, mode := range []string{"denied", "nil-verifier", "cancelled", "already-cancelled"} {
		t.Run(mode, func(t *testing.T) {
			path, capture := portableHelperFixture(t, "echo")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			ready := false
			verify := func(ctx context.Context, _ int) error {
				if err := waitForEmptyCapture(ctx, capture); err != nil {
					return err
				}
				ready = true
				if mode == "cancelled" {
					cancel()
					return nil
				}
				return errors.New("synthetic untrusted process")
			}
			if mode == "nil-verifier" {
				verify = nil
			}
			if mode == "already-cancelled" {
				cancel()
			}
			out, err := runVerifiedHelper(ctx, path, []byte("synthetic-request"), verify)
			if !errors.Is(err, ErrUnavailable) || len(out) != 0 {
				t.Fatalf("unapproved helper was accepted: %v", err)
			}
			if (mode == "denied" || mode == "cancelled") && !ready {
				t.Fatal("did not exercise a started, waiting helper")
			}
			assertNoCapturedRequest(t, capture)
		})
	}
}

func TestPortableHelperBoundsOutputAndRejectsExitFailure(t *testing.T) {
	for _, mode := range []string{"overflow", "failure"} {
		t.Run(mode, func(t *testing.T) {
			path, capture := portableHelperFixture(t, mode)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			input := []byte("synthetic-request")
			out, err := runVerifiedHelper(ctx, path, input, func(ctx context.Context, _ int) error {
				return waitForEmptyCapture(ctx, capture)
			})
			if !errors.Is(err, ErrUnavailable) || len(out) != 0 {
				t.Fatalf("bad helper output accepted: %v", err)
			}
			data, readErr := os.ReadFile(capture)
			if readErr != nil || !bytes.Equal(data, input) {
				t.Fatal("did not exercise the approved child response")
			}
		})
	}
}

func TestPortableHelperRealVerifierRejectsUntrustedProcess(t *testing.T) {
	path, capture := portableHelperFixture(t, "echo")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := runHelper(ctx, path, []byte("synthetic-request"))
	if !errors.Is(err, ErrUnavailable) || len(out) != 0 {
		t.Fatalf("untrusted test executable accepted: %v", err)
	}
	assertNoCapturedRequest(t, capture)
}

func TestPortableVerifiedHelperDeadlineReturnsNoOutput(t *testing.T) {
	path, capture := portableHelperFixture(t, "wait")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	input := []byte("synthetic-request")
	out, err := runVerifiedHelper(ctx, path, input, func(ctx context.Context, _ int) error {
		return waitForEmptyCapture(ctx, capture)
	})
	if !errors.Is(err, context.DeadlineExceeded) || len(out) != 0 {
		t.Fatalf("verified helper deadline = %v, expected local deadline without output", err)
	}
	data, readErr := os.ReadFile(capture)
	if readErr != nil || !bytes.Equal(data, input) {
		t.Fatal("did not exercise timeout after verified request delivery")
	}
}
