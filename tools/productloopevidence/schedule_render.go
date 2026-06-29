package main

import (
	"fmt"
	"strings"
)

func renderSchedule(b *strings.Builder, m manifest, schedule qaScheduleSource) {
	b.WriteString("## Local QA Schedule Closure\n\n")
	fmt.Fprintf(b, "- SSOT: `%s`\n", m.LocalQAScheduleManifest)
	fmt.Fprintf(b, "- ID: `%s`\n", schedule.ID)
	fmt.Fprintf(b, "- Cadence: `%s`\n", schedule.Cadence)
	fmt.Fprintf(b, "- Freshness window: `%s`\n", schedule.FreshnessWindow)
	fmt.Fprintf(b, "- Entrypoint: `%s`\n", schedule.Entrypoint)
	fmt.Fprintf(b, "- Evidence outputs: `%d`\n\n", len(schedule.Evidence))
}
