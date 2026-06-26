package main

import (
	"fmt"
	"strings"
)

func renderDesktopMSIX(m manifest) string {
	return renderDoc("Desktop/MSIX Consumption and CDN Mirror", "../release-artifacts.md", func(b *strings.Builder) {
		b.WriteString("Riido Desktop treats the daemon binary as a helper release artifact:\n\n")
		writeBullets(b, desktopFlow(m))
		fmt.Fprintf(b, "The launcher passes %s when starting the daemon.\n\n", codeList(m.DesktopMSIX.RequiredEnv))
		b.WriteString("## CDN latest mirror\n\n")
		fmt.Fprintf(b, "GitHub Releases remain immutable. `%s` is a mutable development/test mirror only.\n\n", m.DesktopMSIX.CDNLatestBaseURL)
		fmt.Fprintf(b, "The mirror is published by `%s` using `%s` after release artifact CI succeeds.\n\n", m.CDNPublishWorkflow, m.CDNPublishScript)
		b.WriteString("Mirror paths:\n\n")
		writeBullets(b, m.DesktopMSIX.CDNLatestPaths)
		b.WriteString("The archive `VERSION` file must identify the release tag that was mirrored.\n")
	})
}

func desktopFlow(m manifest) []string {
	return []string{
		"select the asset by platform and architecture",
		"download it over HTTPS from " + m.DesktopMSIX.DownloadSource,
		"verify `SHA256SUMS`",
		"extract the binary under " + m.DesktopMSIX.StorageRoot,
		"execute it as the current user without requiring administrator-owned paths",
	}
}
