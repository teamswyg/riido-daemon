package main

import (
	"fmt"
	"strings"
)

func renderAssets(m manifest) string {
	return renderDoc("Decision and Asset Names", "../release-artifacts.md", func(b *strings.Builder) {
		fmt.Fprintf(b, "`riido-daemon` publishes assets from `%s` using `%s`.\n\n", m.ReleaseWorkflow, m.BuildScript)
		b.WriteString("| Platform | Target | Asset |\n| --- | --- | --- |\n")
		for _, target := range m.Targets {
			fmt.Fprintf(b, "| %s | `%s/%s/%s` | `%s` |\n",
				target.Platform, target.GOOS, target.GOARCH, target.Format, assetName(target))
		}
		b.WriteString("\nArchives contain:\n\n")
		writeBullets(b, m.ArchiveContents)
		b.WriteString("Checksums are published as `SHA256SUMS` by the release publish script.\n")
	})
}

func assetName(t target) string {
	ext := ".tar.gz"
	if t.Format == "zip" {
		ext = ".zip"
	}
	return fmt.Sprintf("riido-daemon_%s_%s%s", t.GOOS, t.GOARCH, ext)
}
