package main

import "github.com/teamswyg/riido-daemon/internal/provider/claude"

func claudeModelLabel(id string) string {
	return claude.ModelLabel(id)
}
