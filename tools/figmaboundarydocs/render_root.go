package main

import (
	"fmt"
	"strings"
)

func renderRoot(m manifest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Figma AI Agent Daemon Boundary\n\n%s\n\n", generatedHeader)
	fmt.Fprintf(&b, "Riido task: %s.\n\n", m.RiidoTask)
	b.WriteString("This is the daemon-side projection of Figma AI Agent UI evidence. ")
	b.WriteString("The canonical UI coverage owner remains `riido-contracts`; daemon records ")
	b.WriteString("only executable assignment, runtime, liveness, lifecycle, and provider-input boundaries.\n\n")
	renderRootManifest(&b, m)
	renderRootAnchors(&b, m)
	renderRootSections(&b)
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func renderRootManifest(b *strings.Builder, m manifest) {
	fmt.Fprintf(b, "Executable manifest: [`%s`](figma-ai-agent-daemon-boundary.riido.json), schema `%s`.\n\n",
		"figma-ai-agent-daemon-boundary.riido.json", m.SchemaVersion)
}

func renderRootAnchors(b *strings.Builder, m manifest) {
	p := m.SourceCoverageManifestProvenance
	b.WriteString("Decision anchors:\n\n")
	fmt.Fprintf(b, "- hardening tasks: %s\n", sentenceList(m.HardeningTasks))
	fmt.Fprintf(b, "- upstream provenance: `%s`, `%s`\n", p.Repo, p.ID)
	fmt.Fprintf(b, "- mirrored field: `%s`\n", p.MirrorsSourceField)
	fmt.Fprintf(b, "- source field introduced by: `%s`\n", p.SourceFieldIntroducedBy)
	fmt.Fprintf(b, "- stabilized by: `%s`\n", strings.Join(p.StabilizedBy, "`, `"))
	renderRootLimitations(b, m.MirroredSupportingToolLimitations)
	renderRootEntryAnchors(b, m)
	b.WriteByte('\n')
}

func renderRootLimitations(b *strings.Builder, limitations []toolLimitation) {
	for _, limitation := range limitations {
		fmt.Fprintf(b, "- limitation: `%s`\n", limitation.SourceID)
		fmt.Fprintf(b, "- required authoritative pages: `%s`\n",
			strings.Join(limitation.RequiredAuthoritativePages, "`, `"))
	}
}

func renderRootEntryAnchors(b *strings.Builder, m manifest) {
	fmt.Fprintf(b, "- representative nodes: `%s`, `%s`\n", "432:37336", "432:46849")
	fmt.Fprintf(b, "- draft warning: %s\n", entryScope(m, "432:46849"))
	fmt.Fprintf(b, "- vocabulary warning: fixture is not a daemon template entity\n")
	b.WriteString("- loop direction: Top-down and Bottom-up changes stay explicit\n")
}

func renderRootSections(b *strings.Builder) {
	b.WriteString("Focused sections:\n\n")
	for _, link := range detailLinks() {
		fmt.Fprintf(b, "- [%s](figma-ai-agent-daemon-boundary/%s)\n", link.Title, link.File)
	}
}
