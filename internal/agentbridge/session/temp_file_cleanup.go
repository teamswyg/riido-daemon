package session

import (
	"errors"
	"fmt"
	"os"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func cleanupTempFiles(paths []string) []agentbridge.Event {
	var out []agentbridge.Event
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			out = append(out, agentbridge.Event{
				Kind: agentbridge.EventWarning,
				Text: "adapter temp file cleanup failed",
				Err:  fmt.Sprintf("%s: %v", path, err),
			})
		}
	}
	return out
}
