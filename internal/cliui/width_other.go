//go:build !(darwin || linux || freebsd || netbsd || openbsd)

package cliui

import "os"

// Without a portable size query the caller's COLUMNS or the 80-cell default applies.
func terminalWidth(*os.File) int { return 0 }

// isTerminal falls back to the character-device check where no terminal
// driver query is available.
func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
