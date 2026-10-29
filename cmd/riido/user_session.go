package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

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
	if c.userSession != nil {
		return c.userSession(args)
	}
	origin, _, err := c.userAPIOrigin()
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
	digest := sha256.Sum256([]byte("riido-cli:" + origin))
	native := &securecredential.Store{HelperPath: filepath.Join(filepath.Dir(executable), "RiidoCLIAuthHelper.app", "Contents", "MacOS", "riido-cli-auth-helper"), Key: hex.EncodeToString(digest[:]), Interactive: !args.NoInput, LockDir: filepath.Join(configDir, "riido-cli", "auth-locks")}
	return standalonecli.New(standalonecli.Config{APIOrigin: origin, Store: &standalonecli.ProtectedStore{Native: native, Resource: origin}, OpenBrowser: openUserBrowser})
}

func openUserBrowser(ctx context.Context, address string) error {
	if runtime.GOOS != "darwin" {
		return standalonecli.ErrBrowserUnavailable
	}
	return exec.CommandContext(ctx, "/usr/bin/open", address).Run()
}

// userAPIOrigin validates the same OAuth resource without opening storage or
// contacting the network. Invalid input is never returned to diagnostics.
func (c contextIO) userAPIOrigin() (string, string, error) {
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
