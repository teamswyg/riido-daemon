//go:build windows

package main

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"unsafe"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestWindowsAuthorizationLauncherNativeCallContract(t *testing.T) {
	address := "https://example.test/oauth/authorize?state=a%26b&label=한글"
	want, _ := syscall.UTF16FromString(address)
	for _, mode := range []string{"success", "init-failed", "launch-failed", "cancel-before", "cancel-during-init", "cancel-during-launch"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			loaded, initialized, launched, cleaned := 0, 0, 0, 0
			if mode == "cancel-before" {
				cancel()
			}
			err := openWindowsAuthorizationURI(ctx, address, func() (authorizationShell, error) {
				loaded++
				return authorizationShell{
					initialize: func() bool {
						initialized++
						if mode == "cancel-during-init" {
							cancel()
						}
						return mode != "init-failed"
					},
					uninitialize: func() { cleaned++ },
					execute: func(info *authorizationShellInfo) bool {
						launched++
						if info.Size != uint32(unsafe.Sizeof(*info)) || info.Mask != 0x500 || info.Show != 1 || info.Parameters != nil || info.Dir != nil || info.Process != 0 || info.Window != 0 {
							t.Fatal("native shell options changed")
						}
						got := unsafe.Slice(info.File, len(want))
						for i := range want {
							if got[i] != want[i] {
								t.Fatal("Unicode URI changed")
							}
						}
						if mode == "cancel-during-launch" {
							cancel()
						}
						return mode != "launch-failed"
					},
				}, nil
			})
			switch mode {
			case "success":
				if err != nil || loaded != 1 || initialized != 1 || launched != 1 || cleaned != 1 {
					t.Fatal("successful launch did not clean COM")
				}
			case "init-failed":
				if !errors.Is(err, standalonecli.ErrBrowserUnavailable) || launched != 0 || cleaned != 0 {
					t.Fatal("failed COM initialization used shell")
				}
			case "launch-failed":
				if !errors.Is(err, standalonecli.ErrBrowserUnavailable) || cleaned != 1 {
					t.Fatal("native failure escaped or missed cleanup")
				}
			case "cancel-before":
				if !errors.Is(err, context.Canceled) || loaded != 0 {
					t.Fatal("cancelled request loaded shell")
				}
			case "cancel-during-init":
				if !errors.Is(err, context.Canceled) || launched != 0 || cleaned != 1 {
					t.Fatal("cancelled init launched")
				}
			case "cancel-during-launch":
				if !errors.Is(err, context.Canceled) || cleaned != 1 {
					t.Fatal("late launch accepted cancelled flow")
				}
			}
		})
	}
}

func TestWindowsAuthorizationNativeAvailabilityWithoutLaunching(t *testing.T) {
	// Resolve real System32 exports and exercise COM only. No browser/app opens.
	t.Setenv("WINDIR", t.TempDir())
	dir, err := authorizationSystemDirectory()
	if err != nil || !filepath.IsAbs(dir) {
		t.Fatal("native OS directory unavailable")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ops, err := loadWindowsAuthorizationShell()
	if err != nil || !ops.initialize() {
		t.Fatal("native shell/COM unavailable")
	}
	ops.uninitialize()
	missing := filepath.Join(t.TempDir(), "does-not-exist.exe")
	if err := openWindowsAuthorizationURI(t.Context(), missing, loadWindowsAuthorizationShell); !errors.Is(err, standalonecli.ErrBrowserUnavailable) {
		t.Fatal("native ShellExecute failure did not stay private")
	}
	wantSize := uintptr(112)
	if unsafe.Sizeof(uintptr(0)) == 4 {
		wantSize = 60
	}
	if unsafe.Sizeof(authorizationShellInfo{}) != wantSize {
		t.Fatal("incorrect native ABI layout")
	}
	store := &securecredential.Store{HelperPath: filepath.Join(t.TempDir(), "untrusted.exe"), Key: "test-resource-0123456789", Interactive: true}
	if _, err := store.Status(t.Context()); !errors.Is(err, securecredential.ErrUnavailable) {
		t.Fatal("plumbing enabled unverified credentials")
	}
}
