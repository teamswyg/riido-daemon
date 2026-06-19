package main

import (
	"fmt"
	"strings"
)

func renderInstaller(m manifest) string {
	return renderDoc("Installer", "../release-artifacts.md", func(b *strings.Builder) {
		b.WriteString("macOS/Linux users can install the latest release with:\n\n")
		fmt.Fprintf(b, "```bash\n%s\n```\n\n", m.Installer.Command)
		b.WriteString("The installer script:\n\n")
		writeBullets(b, installerFacts(m))
		b.WriteString("Use a specific version with:\n\n")
		fmt.Fprintf(b, "```bash\n%s=v0.0.1 \\\n%s\n```\n", m.Installer.VersionEnv, m.Installer.Command)
	})
}

func installerFacts(m manifest) []string {
	return []string{
		"detects " + codeList(m.Installer.SupportedGOOS) + " and " + codeList(m.Installer.SupportedGOARCH),
		"resolves latest through the GitHub Releases API so prerelease tags are installable",
		"downloads the matching release asset and `SHA256SUMS`",
		"verifies the checksum before installation",
		fmt.Sprintf("installs `riido` to `%s` unless `%s` is set", m.Installer.DefaultInstallDir, m.Installer.InstallDirEnv),
	}
}
