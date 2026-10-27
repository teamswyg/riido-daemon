//go:build darwin

package securecredential

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"unicode"
)

func verifyHelper(ctx context.Context, path string) error {
	if !filepath.IsAbs(path) || containsControl(path) {
		return ErrUnavailable
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	verify := exec.CommandContext(ctx, "/usr/bin/codesign", "--verify", "--strict", path)
	if verify.Run() != nil {
		return ErrUnavailable
	}
	var metadata limitedOutput
	inspect := exec.CommandContext(ctx, "/usr/bin/codesign", "--display", "--verbose=4", path)
	inspect.Stdout = &metadata
	inspect.Stderr = &metadata
	if inspect.Run() != nil || metadata.overflow {
		return ErrUnavailable
	}
	var entitlements limitedOutput
	inspectEntitlements := exec.CommandContext(ctx, "/usr/bin/codesign", "--display", "--entitlements", ":-", "--xml", path)
	inspectEntitlements.Stdout = &entitlements
	if inspectEntitlements.Run() != nil || entitlements.overflow {
		return ErrUnavailable
	}
	return checkSignature(metadata.String(), entitlements.Bytes())
}

func containsControl(path string) bool {
	for _, r := range path {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}
