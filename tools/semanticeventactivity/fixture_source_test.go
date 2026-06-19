package main

import (
	"sort"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func eventKindSource() string {
	kinds := make([]string, 0, len(agentbridge.EventKinds()))
	for _, kind := range agentbridge.EventKinds() {
		kinds = append(kinds, string(kind))
	}
	sort.Strings(kinds)
	var b strings.Builder
	b.WriteString("package agentbridge\n\ntype EventKind string\n\nconst (\n")
	for _, kind := range kinds {
		b.WriteString("\tEvent")
		b.WriteString(strings.ReplaceAll(kind, "_", ""))
		b.WriteString(" EventKind = \"")
		b.WriteString(kind)
		b.WriteString("\"\n")
	}
	b.WriteString(")\n")
	return b.String()
}
