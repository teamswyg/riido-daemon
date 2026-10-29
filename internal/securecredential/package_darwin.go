//go:build darwin

package securecredential

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// VerifyPackagedHelper checks a copied, pre-signed bundle without executing it
// or accessing credentials. Each architecture's hardened runtime and entitlements
// must satisfy the same pinned identity policy as the credential helper runtime.
func VerifyPackagedHelper(ctx context.Context, bundle string, architectures []string) error {
	if !filepath.IsAbs(bundle) || containsControl(bundle) || len(architectures) == 0 {
		return ErrUnavailable
	}
	info, err := os.Lstat(bundle)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return ErrUnavailable
	}
	verifyCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if exec.CommandContext(verifyCtx, "/usr/bin/codesign", "--verify", "--deep", "--strict", "--all-architectures", "-R="+helperRequirement, bundle).Run() != nil {
		return ErrUnavailable
	}
	for _, arch := range architectures {
		if arch != "arm64" && arch != "x86_64" {
			return ErrUnavailable
		}
		if err := verifyCodeTarget(ctx, bundle, "--architecture", arch); err != nil {
			return err
		}
	}
	return nil
}
