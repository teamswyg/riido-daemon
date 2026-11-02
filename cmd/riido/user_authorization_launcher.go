package main

import (
	"context"
	"net"
	"net/url"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/standalonecli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type authorizationLauncher func(context.Context, string) error

func openUserBrowserWith(ctx context.Context, address string, launch authorizationLauncher) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	u, err := url.Parse(address)
	if err != nil || len(address) > 16<<10 || strings.ContainsAny(address, "\x00\r\n\t \"<>") || u.Host == "" || u.User != nil || u.Fragment != "" || u.Opaque != "" || u.String() != address {
		return standalonecli.ErrBrowserUnavailable
	}
	ip := net.ParseIP(u.Hostname())
	if u.Scheme != "https" && (u.Scheme != "http" || ip == nil || !ip.IsLoopback()) {
		return standalonecli.ErrBrowserUnavailable
	}
	return launchAuthorizationURI(ctx, address, launch, standalonecli.ErrBrowserUnavailable)
}

func openUserDesktopWith(ctx context.Context, request userauth.DesktopAuthorizationRequest, launch authorizationLauncher) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	address, err := request.LaunchURL()
	if err != nil {
		return standalonecli.ErrDesktopUnavailable
	}
	return launchAuthorizationURI(ctx, address, launch, standalonecli.ErrDesktopUnavailable)
}

func launchAuthorizationURI(ctx context.Context, address string, launch authorizationLauncher, unavailable error) error {
	if launch == nil {
		return unavailable
	}
	err := launch(ctx, address)
	if cancelled := ctx.Err(); cancelled != nil {
		return cancelled
	}
	if err != nil {
		return unavailable
	}
	return nil
}
