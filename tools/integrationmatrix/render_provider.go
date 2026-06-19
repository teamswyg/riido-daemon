package main

import (
	"fmt"
	"strings"
)

func renderProviderMatrix(m manifest) string {
	return renderDoc("Provider Matrix", "../integration-matrix.md", func(b *strings.Builder) {
		b.WriteString("| Provider | Executable | Public deterministic CI | Real CLI integration | Worktree routing status |\n")
		b.WriteString("| --- | --- | --- | --- | --- |\n")
		for _, p := range m.ProviderValidation.Providers {
			fmt.Fprintf(b, "| %s | `%s` | %s | `%s` | %s |\n",
				p.DisplayName, executableLabel(m, p), codeList(p.DeterministicCI), p.OptInIntegration, worktreeStatus(p))
		}
		b.WriteString("\nPublic pull-request CI runs deterministic tests and keeps real provider integration opt-in.\n")
		b.WriteString("The scheduled/manual [`provider-real-cli-observation`](../provider-real-cli-observation.md) workflow records available-provider evidence.\n\n")
		b.WriteString("Operators can run the current local provider matrix with:\n\n```bash\n./scripts/integration-smoke.sh\n```\n")
	})
}

func executableLabel(m manifest, p providerEvidence) string {
	if p.Provider == "codex" {
		return "codex --sandbox danger-full-access app-server --listen stdio://"
	}
	if found, ok := realProviderByID(m, p.Provider); ok {
		return found.DefaultExecutable
	}
	return p.Executable
}
