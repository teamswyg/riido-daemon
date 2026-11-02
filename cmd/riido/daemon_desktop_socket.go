package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/teamswyg/riido-daemon/internal/cliconfig"
)

// desktopDaemonSocket finds the daemon that the Riido desktop app runs for the
// build's default profile. The app keeps one daemon per profile in its own
// data folder, so a standalone CLI would otherwise only try the developer
// default socket and report the running app daemon as unreachable.
func desktopDaemonSocket() (string, bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}
	return desktopDaemonSocketFor(runtime.GOOS, home, cliconfig.DefaultProfile(), isUnixSocket)
}

func desktopDaemonSocketFor(goos, home, profile string, exists func(string) bool) (string, bool) {
	if goos != "darwin" || home == "" {
		return "", false
	}
	dir := "ai-agent-daemon"
	switch profile {
	case "production":
	case "staging":
		dir += "-staging"
	default:
		return "", false
	}
	socket := filepath.Join(home, "Library", "Application Support", "riido-desktop", dir, "agentd.sock")
	return socket, exists(socket)
}

func isUnixSocket(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode()&os.ModeSocket != 0
}
