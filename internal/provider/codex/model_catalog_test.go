package codex

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func catalogFixture(t *testing.T, script string) agentbridge.DetectEnv {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("Unix shell fixture")
	}
	exe := filepath.Join(t.TempDir(), "codex")
	if err := os.WriteFile(exe, []byte("#!/bin/sh\n"+script), 0o700); err != nil {
		t.Fatal(err)
	}
	return agentbridge.DetectEnv{EnvOverride: map[string]string{EnvOverride: exe}}
}

func TestCatalogUsesNativeModelsAndPaginates(t *testing.T) {
	env := catalogFixture(t, `read initialize
printf '%s\n' '{"id":1,"result":{}}'
read initialized
read first
printf '%s\n' '{"id":2,"result":{"data":[{"id":"ui-id","model":"native-model","displayName":"Native","isDefault":true}],"nextCursor":"page-2"}}'
read second
case "$second" in *page-2*) ;; *) exit 1;; esac
printf '%s\n' '{"id":3,"result":{"data":[{"id":"other-model"}],"nextCursor":null}}'
`)
	models, err := ModelCatalog(context.Background(), env)
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 2 || models[0].ModelID != "native-model" || models[0].Label != "Native" || !models[0].IsDefault || models[1].ModelID != "other-model" {
		t.Fatalf("catalog=%+v", models)
	}
}

func TestCatalogProbeFailsBoundedlyWithoutInventingModels(t *testing.T) {
	for _, script := range []string{"exec sleep 30\n", "read initialize\nprintf '%s\\n' '{\"id\":1,\"error\":{\"message\":\"unsupported\"}}'\n"} {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		start := time.Now()
		models, err := ModelCatalog(ctx, catalogFixture(t, script))
		cancel()
		if err == nil || len(models) != 0 || time.Since(start) > 2*time.Second {
			t.Fatalf("models=%+v err=%v elapsed=%v", models, err, time.Since(start))
		}
	}
}
