package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/cliconfig"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

const (
	cliDefaultAPIOrigin = "https://monolith.riido.io"
	cliAPIOriginEnv     = "RIIDO_CLI_API_URL"
)

func (c contextIO) openUserSession(args contextcli.Args) (*standalonecli.Service, error) {
	if args.SettingsResolved && args.APIOrigin == "" {
		return nil, cliconfig.ErrInvalid
	}
	authOrigin := args.AuthOrigin
	if authOrigin == "" && !args.SettingsResolved {
		authOrigin = userauth.AuthOriginBrowser
	}
	if !cliconfig.ValidAuthOrigin(authOrigin) {
		return nil, cliconfig.ErrInvalid
	}
	if c.userSession != nil {
		return c.userSession(args)
	}
	origin, _, err := c.userAPIOrigin(args)
	if err != nil {
		return nil, err
	}
	executable, err := c.executable()
	if err != nil {
		return nil, securecredential.ErrUnavailable
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return nil, securecredential.ErrUnavailable
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, securecredential.ErrUnavailable
	}
	native := securecredential.Store{HelperPath: filepath.Join(filepath.Dir(executable), "RiidoCLIAuthHelper.app", "Contents", "MacOS", "riido-cli-auth-helper"), Interactive: !args.NoInput, LockDir: filepath.Join(configDir, "riido-cli", "auth-locks")}
	store, err := standalonecli.NewProtectedStore(native, origin, authOrigin)
	if err != nil {
		return nil, err
	}
	return standalonecli.New(standalonecli.Config{APIOrigin: origin, AuthOrigin: authOrigin, Store: store, OpenBrowser: openUserBrowser})
}

func openUserBrowser(ctx context.Context, address string) error {
	if runtime.GOOS != "darwin" {
		return standalonecli.ErrBrowserUnavailable
	}
	return exec.CommandContext(ctx, "/usr/bin/open", address).Run()
}

// userAPIOrigin validates the same OAuth resource without opening storage or
// contacting the network. Invalid input is never returned to diagnostics.
func (c contextIO) userAPIOrigin(args ...contextcli.Args) (string, string, error) {
	if len(args) > 0 && args[0].SettingsResolved {
		if args[0].APIOrigin == "" {
			return "", args[0].APIOriginSource, cliconfig.ErrInvalid
		}
		return args[0].APIOrigin, args[0].APIOriginSource, nil
	}
	origin := strings.TrimSpace(c.getenv(cliAPIOriginEnv))
	source := "environment"
	if origin == "" {
		origin = cliDefaultAPIOrigin
		source = "default"
	}
	origin = strings.TrimRight(origin, "/")
	_, err := userauth.NewClient(userauth.Config{ClientID: "riido-cli", Resource: origin, Scopes: []string{"read"}, Endpoints: userauth.Endpoints{AuthorizeURL: origin + "/oauth/authorize", TokenURL: origin + "/oauth/token", RevokeURL: origin + "/oauth/revoke"}})
	if err != nil {
		return "", source, err
	}
	return origin, source, nil
}
