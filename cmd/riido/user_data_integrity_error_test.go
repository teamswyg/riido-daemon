package main

import (
	"fmt"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

// RIID-6008: broken stored data has its own code and exit status, distinct from
// permission (4), unavailable (6) and version conflict (10).
func TestUserDataIntegrityErrorHasOwnExit(t *testing.T) {
	got := userAuthError(fmt.Errorf("read: %w", userapi.ErrDataIntegrity))
	if got.Code != contextcli.CodeDataIntegrity || got.ExitCode() != contextcli.ExitDataIntegrity || got.ExitCode() != 11 || got.Message == "" || got.NextAction == "" {
		t.Fatalf("integrity mapping = %+v", got)
	}
	if conflict := userAuthError(userapi.ErrConflict); conflict.Code != contextcli.CodeConflict {
		t.Fatal("version conflict mapping changed", conflict.Code)
	}
}
