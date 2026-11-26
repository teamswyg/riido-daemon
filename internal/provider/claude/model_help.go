package claude

import "strings"

// ModelLabel returns the catalog display name for a Claude Code model ID.
// CLI help lists examples rather than a complete map of alias versions.
func ModelLabel(id string) string {
	switch id {
	case "opus":
		return "Opus 4.8"
	case "sonnet":
		return "Sonnet 5"
	case "haiku":
		return "Haiku 4.5"
	case "fable":
		return "Fable 5"
	}

	parts := strings.Split(id, "-")
	for i, part := range parts {
		if part != "" {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

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
