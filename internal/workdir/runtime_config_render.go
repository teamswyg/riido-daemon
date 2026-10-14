package workdir

import (
	"fmt"
	"strings"
)

func renderRuntimeConfig(cfg RuntimeConfig) string {
	var b strings.Builder
	b.WriteString("# Runtime configuration\n\n")
	if cfg.Identity != "" {
		b.WriteString("## Identity\n\n")
		b.WriteString(cfg.Identity)
		b.WriteString("\n\n")
	}
	if len(cfg.CLICatalog) > 0 {
		b.WriteString("## CLI catalog\n\n")
		for _, line := range cfg.CLICatalog {
			b.WriteString("- `")
			b.WriteString(line)
			b.WriteString("`\n")
		}
		b.WriteString("\n")
	}
	if len(cfg.HardRules) > 0 {
		b.WriteString("## Hard rules\n\n")
		for _, r := range cfg.HardRules {
			b.WriteString("- ")
			b.WriteString(r)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	workflow := cfg.Workflow
	if workflow == "" {
		workflow = "default"
	}
	fmt.Fprintf(&b, "## Workflow\n\nworkflow: %s\n", workflow)
	return b.String()
}
