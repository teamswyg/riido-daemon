package main

import (
	"fmt"
	"strings"
)

func renderDetailBullets(b *strings.Builder, items []string) {
	for _, item := range items {
		fmt.Fprintf(b, "- %s\n", item)
	}
	b.WriteByte('\n')
}

func renderCommandFence(b *strings.Builder, items []string) {
	b.WriteString("```bash\n")
	for _, item := range items {
		fmt.Fprintf(b, "%s\n", item)
	}
	b.WriteString("```\n\n")
}

func renderSimpleCommandGroups(b *strings.Builder, groups []CommandGroup) {
	b.WriteString("| Command group | Backing owner | Boundary |\n| --- | --- | --- |\n")
	for _, group := range groups {
		fmt.Fprintf(b, "| `riido %s%s` | %s | %s |\n", group.Name, subcommandEllipsis(group), group.Owner, group.Boundary)
	}
	b.WriteByte('\n')
}

func subcommandEllipsis(group CommandGroup) string {
	if len(group.Subcommands) == 0 {
		return ""
	}
	return " ..."
}
