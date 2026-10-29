// clihelperpackage prepares signed optional helpers or checks portable archive structure.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
)

func main() {
	bundle := flag.String("bundle", "", "pre-signed helper bundle (macOS only)")
	output := flag.String("out", "", "fresh package directory")
	archive := flag.String("archive", "", "archive to structurally verify without executing it")
	version := flag.String("version", "", "exact release version")
	arch := flag.String("arch", "", "Go target architecture")
	recheck := flag.Bool("recheck-signature", false, "unpack and verify helper signature on macOS")
	flag.Parse()
	var err error
	switch {
	case *archive != "" && *bundle == "" && *output == "":
		err = verifyArchive(*archive, *version, *arch)
		if err == nil && *recheck {
			err = recheckArchive(context.Background(), *archive, *version, *arch, securecredential.VerifyPackagedHelper)
		}
	case *bundle != "" && *archive == "" && *output != "" && !*recheck:
		err = prepareBundle(context.Background(), *bundle, *output, *version, *arch, securecredential.VerifyPackagedHelper)
	default:
		err = fmt.Errorf("provide either -bundle/-out or -archive")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "CLI helper package:", err)
		os.Exit(1)
	}
}
