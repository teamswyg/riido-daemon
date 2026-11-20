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
		if m.Installer.ChannelEnv != "" {
			b.WriteString("\nInstall the package the desktop app uses for a channel (" + codeList(m.Installer.Channels) + ") from the CDN with:\n\n")
			fmt.Fprintf(b, "```bash\ncurl -fsSL %s | %s=staging sh\n```\n", installerScriptURL(m), m.Installer.ChannelEnv)
			if len(m.Installer.ChannelNotes) > 0 {
				b.WriteString("\n")
				writeBullets(b, m.Installer.ChannelNotes)
			}
		}
	})
}

func installerFacts(m manifest) []string {
	return []string{
		"detects " + codeList(m.Installer.SupportedGOOS) + " and " + codeList(m.Installer.SupportedGOARCH),
		"resolves latest through the GitHub Releases API so prerelease tags are installable",
		"downloads the matching release asset and `SHA256SUMS`",
		"verifies the checksum before installation",
		fmt.Sprintf("installs `riido` to `%s` unless `%s` is set", m.Installer.DefaultInstallDir, m.Installer.InstallDirEnv),
		"installs the adjacent `VERSION`, and on macOS the signed `RiidoCLIAuthHelper.app` with `CLI_AUTH_HELPER.json`; a package without the helper removes an older helper, and a package with only one of the pair is rejected before the existing binary is touched",
		"prints the PATH line for the current shell, or warns when another `riido` is found first on PATH",
	}
}

// installerScriptURL is the script URL part of the documented curl command.
func installerScriptURL(m manifest) string {
	for field := range strings.FieldsSeq(m.Installer.Command) {
		if strings.HasPrefix(field, "https://") {
			return field
		}
	}
	return m.Installer.Command
}
