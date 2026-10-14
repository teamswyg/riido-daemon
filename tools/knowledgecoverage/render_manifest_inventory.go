package main

import (
	"fmt"
	"strings"
)

func renderManifestInventory(b *strings.Builder, inventory manifestInventory) {
	b.WriteString("## Manifest Inventory\n\n")
	b.WriteString("| Group | Count | Sample paths |\n| --- | ---: | --- |\n")
	for _, group := range inventory.Groups {
		fmt.Fprintf(b, "| `%s` | %d | %s |\n", group.Group, group.Count, manifestSampleText(inventory, group.Group))
	}
	b.WriteString("\n")
}

func manifestSampleText(inventory manifestInventory, group string) string {
	for _, sample := range inventory.Samples {
		if sample.Group == group {
			return samplePathText(sample.Paths)
		}
	}
	return "None"
}

func samplePathText(paths []string) string {
	if len(paths) == 0 {
		return "None"
	}
	quoted := make([]string, 0, len(paths))
	for _, path := range paths {
		quoted = append(quoted, "`"+path+"`")
	}
	return strings.Join(quoted, "<br>")
}
