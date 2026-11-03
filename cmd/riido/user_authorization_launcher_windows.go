//go:build windows

package main

import (
	"context"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

// Shell extensions may require STA. Keep initialization, launch and cleanup on
// one OS thread; URL dispatch is never a command line interpreted by a shell.
func openUserAuthorizationURI(ctx context.Context, address string) error {
	return openWindowsAuthorizationURI(ctx, address, loadWindowsAuthorizationShell)
}

func openWindowsAuthorizationURI(ctx context.Context, address string, load func() (authorizationShell, error)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	file, err := syscall.UTF16PtrFromString(address)
	if err != nil || address == "" {
		return standalonecli.ErrBrowserUnavailable
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ops, err := load()
	if err != nil || ops.initialize == nil || ops.uninitialize == nil || ops.execute == nil {
		return standalonecli.ErrBrowserUnavailable
	}
	if !ops.initialize() {
		return standalonecli.ErrBrowserUnavailable
	}
	defer ops.uninitialize()
	if err := ctx.Err(); err != nil {
		return err
	}
	verb, _ := syscall.UTF16PtrFromString("open")
	info := authorizationShellInfo{Mask: 0x100 | 0x400, Verb: verb, File: file, Show: 1}
	info.Size = uint32(unsafe.Sizeof(info))
	ok := ops.execute(&info)
	runtime.KeepAlive(file)
	runtime.KeepAlive(verb)
	if err := ctx.Err(); err != nil {
		return err
	}
	if !ok {
		return standalonecli.ErrBrowserUnavailable
	}
	return nil
}

type authorizationShell struct {
	initialize   func() bool
	uninitialize func()
	execute      func(*authorizationShellInfo) bool
}

// SHELLEXECUTEINFOW has pointer-sized handles and platform-native alignment.
// No parameters, working directory, parent window or process handle is supplied.
type authorizationShellInfo struct {
	Size, Mask       uint32
	Window           uintptr
	Verb, File       *uint16
	Parameters, Dir  *uint16
	Show             int32
	Instance, IDList uintptr
	Class            *uint16
	ClassKey         uintptr
	HotKey           uint32
	Icon, Process    uintptr
}
