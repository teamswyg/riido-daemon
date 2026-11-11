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
	"github.com/teamswyg/riido-daemon/internal/userauth"
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
	state, err := saasplane.DeviceBootstrap(ctx, args[0], raw, &securecredential.Store{HelperPath: helper, LockDir: locks, Interactive: true, SharedSession: true}, defaultContextIO().nativeBootstrapAdmission)
	if err != nil {
		return fail
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		State string `json:"state"`
	}{state})
}

// The native process resolves the same managed profile as Desktop. The peer's
// claimed expiry alone can never authorize renewal or a first prepare.
func (c contextIO) nativeBootstrapAdmission(ctx context.Context, binding userauth.BootstrapRequest) (userauth.DeviceLinkApproval, error) {
	var empty userauth.DeviceLinkApproval
	var profile string
	switch binding.Resource {
	case "https://monolith.api.riido.io":
		profile = "staging"
	case "https://monolith.riido.io":
		profile = "production"
	default:
		return empty, errors.New("invalid bootstrap resource")
	}
	args, cliErr := c.parseArgs([]string{"--profile", profile, "--json"}, contextcli.ArgSpec{})
	if cliErr != nil || args.APIOrigin != binding.Resource {
		return empty, errors.New("invalid bootstrap profile")
	}
	service, err := c.openUserSession(args)
	if err != nil {
		return empty, err
	}
	return service.BoundBootstrapRequest(ctx, binding)
}
