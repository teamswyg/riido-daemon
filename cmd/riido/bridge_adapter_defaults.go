package main

import (
	"github.com/teamswyg/riido-daemon/internal/provider/claude"
	"github.com/teamswyg/riido-daemon/internal/provider/codex"
	"github.com/teamswyg/riido-daemon/internal/provider/cursor"
	"github.com/teamswyg/riido-daemon/internal/provider/openclaw"
)

// providerDefaultExecutable returns the binary name an adapter looks up on
// $PATH when no explicit override is given.
func providerDefaultExecutable(name string) string {
	switch name {
	case claude.Name:
		return claude.DefaultExecutable
	case codex.Name:
		return codex.DefaultExecutable
	case openclaw.Name:
		return openclaw.DefaultExecutable
	case cursor.Name:
		return cursor.DefaultExecutable
	}
	return ""
}
