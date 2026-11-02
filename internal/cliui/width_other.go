//go:build !(darwin || linux || freebsd || netbsd || openbsd)

package cliui

import "os"

// Without a portable size query the caller's COLUMNS or the 80-cell default applies.
func terminalWidth(*os.File) int { return 0 }
