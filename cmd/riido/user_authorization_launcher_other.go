//go:build !darwin && !windows

package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func openUserAuthorizationURI(context.Context, string) error {
	return standalonecli.ErrBrowserUnavailable
}
