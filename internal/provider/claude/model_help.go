package claude

import "strings"

// ModelIDsFromHelp combines the built-in aliases with examples advertised by the CLI.
// Claude Code help shows examples and can omit aliases such as Haiku.
func ModelIDsFromHelp(help string) []string {
	_, section, ok := strings.Cut(help, "--model <model>")
	if !ok {
		return nil
	}
	section, _, _ = strings.Cut(section, "\n  -n, --name")
	ids := []string{"opus", "sonnet", "haiku"}
	seen := map[string]bool{"opus": true, "sonnet": true, "haiku": true}
	for token := range strings.SplitSeq(section, "'") {
		id := strings.TrimSpace(token)
		if seen[id] || (id != "fable" && !strings.HasPrefix(id, "claude-")) {
			continue
		}
		ids = append(ids, id)
		seen[id] = true
	}
	return ids
}
