package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane/saasplane"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/securecredential"
)

func runDaemonDeviceBootstrap(ctx context.Context, args []string) error {
	fail := errors.New("device bootstrap could not be confirmed")
	if runtime.GOOS != "darwin" || len(args) != 1 || args[0] != "status" && args[0] != "prepare" && args[0] != "activate" || defaultContextIO().environment().Mode == contextcli.ModeAgent {
		return fail
	}
	raw, err := io.ReadAll(io.LimitReader(os.Stdin, 8193))
	if err != nil || len(raw) > 8192 {
		return fail
	}
	exe, err := os.Executable()
	if err != nil {
		return fail
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fail
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return fail
	}
	helper, locks, err := userSessionPathsForOS(runtime.GOOS, exe, dir, "")
	if err != nil {
		return fail
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	state, err := saasplane.DeviceBootstrap(ctx, args[0], raw, &securecredential.Store{HelperPath: helper, LockDir: locks, Interactive: true, SharedSession: true})
	if err != nil {
		return fail
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		State string `json:"state"`
	}{state})
}
