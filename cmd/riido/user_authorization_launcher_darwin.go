//go:build darwin

package main

import (
	"context"
	"os/exec"
)

func openUserAuthorizationURI(ctx context.Context, address string) error {
	return exec.CommandContext(ctx, "/usr/bin/open", address).Run()
}
