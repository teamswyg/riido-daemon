//go:build darwin || linux

package securecredential

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestSwappedHelperNeverReceivesCredentialInput(t *testing.T) {
	dir := t.TempDir()
	candidate := filepath.Join(dir, "helper")
	replacement := filepath.Join(dir, "replacement")
	capture := filepath.Join(dir, "captured")
	t.Setenv("RIIDO_AUTH_FIXTURE_CAPTURE", capture)
	if err := os.WriteFile(candidate, []byte("#!/bin/sh\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	program := `#!/bin/sh
cat > "$RIIDO_AUTH_FIXTURE_CAPTURE"
printf '%s' '{"version":1,"ok":true,"data":{"present":true}}'
`
	if err := os.WriteFile(replacement, []byte(program), 0o700); err != nil {
		t.Fatal(err)
	}
	store := Store{HelperPath: candidate, Key: "synthetic-fixture-key", verify: func(context.Context, string) error { return os.Rename(replacement, candidate) }}
	err := store.Save(context.Background(), []byte("synthetic-fixture-credential"))
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("replacement accepted: %v", err)
	}
	data, readErr := os.ReadFile(capture)
	if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
		t.Fatal(readErr)
	}
	if len(data) != 0 {
		t.Fatal("unverified child received credential bytes")
	}
}

func TestVerifiedChildReceivesInputOnlyAfterVerifier(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "helper")
	capture := filepath.Join(dir, "captured")
	t.Setenv("RIIDO_AUTH_FIXTURE_CAPTURE", capture)
	program := `#!/bin/sh
: > "$RIIDO_AUTH_FIXTURE_CAPTURE"
cat > "$RIIDO_AUTH_FIXTURE_CAPTURE"
printf '%s' 'fixture-result'
`
	if err := os.WriteFile(path, []byte(program), 0o700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	checked := false
	output, err := runVerifiedHelper(ctx, path, []byte("synthetic-input"), func(ctx context.Context, pid int) error {
		if pid <= 0 {
			t.Fatal("no child identity")
		}
		for {
			data, err := os.ReadFile(capture)
			if err == nil {
				if len(data) != 0 {
					t.Fatal("credential bytes preceded process verification")
				}
				checked = true
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Millisecond):
			}
		}
	})
	if err != nil || !checked || string(output) != "fixture-result" {
		t.Fatalf("verified fixture failed: %v", err)
	}
	data, err := os.ReadFile(capture)
	if err != nil || !bytes.Equal(data, []byte("synthetic-input")) {
		t.Fatal("verified child did not receive exact input")
	}
}

func TestCancelledVerificationReapsChildWithoutInput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var pid int
	_, err := runVerifiedHelper(ctx, "/bin/cat", []byte("synthetic-input"), func(_ context.Context, child int) error {
		pid = child
		cancel()
		return nil
	})
	if !errors.Is(err, ErrUnavailable) || pid == 0 {
		t.Fatalf("cancelled verifier: %v", err)
	}
	if syscall.Kill(pid, 0) == nil {
		t.Fatal("cancelled child was not reaped")
	}
}
