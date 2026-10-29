//go:build darwin

package securecredential

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
	"unicode"
)

// The trusted system codesign command evaluates this requirement against the
// actual running code, not a helper-controlled self-report.
const helperRequirement = `anchor apple generic and identifier "io.riido.cli-auth-helper" and certificate leaf[subject.OU] = "3PVV8DQPL6"`

func verifyHelper(ctx context.Context, path string) error {
	if !filepath.IsAbs(path) || containsControl(path) {
		return ErrUnavailable
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return ErrUnavailable
	}
	return verifyCodeTarget(ctx, path)
}

func verifyHelperProcess(ctx context.Context, pid int) error {
	if pid <= 0 {
		return ErrUnavailable
	}
	return verifyCodeTarget(ctx, "+"+strconv.Itoa(pid))
}

func verifyCodeTarget(ctx context.Context, target string, architecture ...string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	verifyArgs := append([]string{"--verify", "--strict", "--verbose=1", "-R=" + helperRequirement}, architecture...)
	verify := exec.CommandContext(ctx, "/usr/bin/codesign", append(verifyArgs, target)...)
	if verify.Run() != nil {
		return ErrUnavailable
	}
	var metadata limitedOutput
	displayArgs := append([]string{"--display", "--verbose=4"}, architecture...)
	inspect := exec.CommandContext(ctx, "/usr/bin/codesign", append(displayArgs, target)...)
	inspect.Stdout = &metadata
	inspect.Stderr = &metadata
	if inspect.Run() != nil || metadata.overflow {
		return ErrUnavailable
	}
	var entitlements limitedOutput
	entitlementArgs := append([]string{"--display", "--entitlements", ":-", "--xml"}, architecture...)
	inspectEntitlements := exec.CommandContext(ctx, "/usr/bin/codesign", append(entitlementArgs, target)...)
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
