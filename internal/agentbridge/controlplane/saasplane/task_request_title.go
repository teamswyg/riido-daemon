package saasplane

import "strings"

// Read the title from the server's canonical assignment snapshot, before any
// provider instruction injection. It is display metadata, never an instruction.
func assignmentTaskTitle(prompt string) string {
	if !strings.HasPrefix(strings.TrimSpace(prompt), "# Riido AI Agent Assignment") {
		return ""
	}
	inTask := false
	for line := range strings.SplitSeq(prompt, "\n") {
		if line == "## Task" {
			inTask = true
			continue
		}
		if !inTask {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			break
		}
		if title, ok := strings.CutPrefix(line, "- title: "); ok {
			return strings.TrimSpace(title)
		}
	}
	return ""
}
