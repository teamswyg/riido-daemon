package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/provider/cursor"
)

func TestCursorObservedRecordsAuthPreflightWithoutSecret(t *testing.T) {
	exe := writeCursorAboutShim(t, "Not logged in")
	t.Setenv(cursor.APIKeyEnv, "secret-value-must-not-leak")
	observed := cursorObserved(exe)
	auth, ok := observed["auth_preflight"].(cursorAuthPreflight)
	if !ok {
		t.Fatalf("auth_preflight missing: %+v", observed)
	}
	if auth.InteractiveLoggedIn || !auth.HeadlessAPIKeyPresent {
		t.Fatalf("auth=%+v", auth)
	}
	if auth.HeadlessAPIKeyEnv != cursor.APIKeyEnv {
		t.Fatalf("auth=%+v", auth)
	}
}

func TestCursorInteractiveLoggedIn(t *testing.T) {
	if cursorInteractiveLoggedIn(writeCursorAboutShim(t, "Logged in as tester")) != true {
		t.Fatal("expected logged in probe")
	}
	if cursorInteractiveLoggedIn(writeCursorAboutShim(t, "Not logged in")) != false {
		t.Fatal("expected not logged in probe")
	}
	if cursorInteractiveLoggedIn(writeCursorFailingShim(t)) != false {
		t.Fatal("expected failed probe to stay unauthenticated")
	}
}

func writeCursorAboutShim(t *testing.T, output string) string {
	t.Helper()
	exe := filepath.Join(t.TempDir(), "cursor-agent")
	body := "#!/bin/sh\nif [ \"$1\" = about ]; then echo '" + output + "'; fi\n"
	if err := os.WriteFile(exe, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return exe
}

func writeCursorFailingShim(t *testing.T) string {
	t.Helper()
	exe := filepath.Join(t.TempDir(), "cursor-agent")
	if err := os.WriteFile(exe, []byte("#!/bin/sh\nexit 2\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return exe
}
