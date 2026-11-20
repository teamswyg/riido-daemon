// clihelperpackage prepares signed optional helpers or checks portable archive structure.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
)

func main() {
	bundle := flag.String("bundle", "", "pre-signed helper bundle (macOS only)")
	output := flag.String("out", "", "fresh package directory")
	archive := flag.String("archive", "", "archive to structurally verify without executing it")
	version := flag.String("version", "", "exact release version")
	arch := flag.String("arch", "", "Go target architecture")
	recheck := flag.Bool("recheck-signature", false, "unpack and verify helper signature on macOS")
	requireHelper := flag.Bool("require-helper", false, "reject legacy archives without a helper")
	flag.Parse()
	var err error
	switch {
	case *archive != "" && *bundle == "" && *output == "":
		err = verifyArchivePolicy(*archive, *version, *arch, *requireHelper)
		if err == nil && *recheck {
			err = recheckArchive(context.Background(), *archive, *version, *arch, verifyRunnablePackagedHelper)
		}
	case *bundle != "" && *archive == "" && *output != "" && !*recheck:
		err = prepareBundle(context.Background(), *bundle, *output, *version, *arch, verifyRunnablePackagedHelper)
	default:
		err = fmt.Errorf("provide either -bundle/-out or -archive")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "CLI helper package:", err)
		os.Exit(1)
	}
}

// Run from the repository root, like the release scripts. Static Apple identity
// checks precede execution; the probe has no credential or service operations.
func verifyRunnablePackagedHelper(ctx context.Context, bundle string, architectures []string) error {
	if err := securecredential.VerifyPackagedHelper(ctx, bundle, architectures); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "python3", "scripts/verify-macos-helper.py", "verify-bundle", "--bundle", bundle)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("helper profile/startup verification failed: %w", err)
	}
	return nil
}
