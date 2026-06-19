package main

import (
	"fmt"
	"strings"
)

func acceptanceDoc(c contract, id string) string {
	item := channelByID(c, id)
	return renderDoc("Store Distribution Architecture: Mac App Store Acceptance", "architecture", func(b *strings.Builder) {
		b.WriteString("Policy can change; update this executable architecture SSOT and the C11 distribution SSOT together.\n\n")
		fmt.Fprintf(b, "Channel: `%s`\n\n", item.ID)
		writeChannelSummary(b, item)
		b.WriteString("\nRequired surfaces:\n\n")
		writeBullets(b, item.RequiredSurfaces)
		b.WriteString("\nForbidden surfaces:\n\n")
		writeBullets(b, item.ForbiddenSurfaces)
		b.WriteString("\nExternal provider CLI execution requires user-selected/sandbox/security-scoped OS grant 와 App Review approval.\n")
	})
}

func msixAcceptanceDoc(c contract) string {
	return renderDoc("Store Distribution Architecture: MSIX Acceptance", "architecture", func(b *strings.Builder) {
		for _, id := range []string{"msix-sideload", "msix-store"} {
			item := channelByID(c, id)
			fmt.Fprintf(b, "## `%s`\n\n", item.ID)
			writeChannelSummary(b, item)
			b.WriteString("\nRequired surfaces:\n\n")
			writeBullets(b, item.RequiredSurfaces)
			b.WriteString("\nForbidden surfaces:\n\n")
			writeBullets(b, item.ForbiddenSurfaces)
			b.WriteString("\n")
		}
	})
}

func writeChannelSummary(b *strings.Builder, item channel) {
	fmt.Fprintf(b, "- Status: `%s`\n", item.Status)
	fmt.Fprintf(b, "- Runtime role: `%s`\n", item.RuntimeRole)
	fmt.Fprintf(b, "- Background rule: `%s`\n", item.BackgroundRule)
	fmt.Fprintf(b, "- IPC/data root: `%s` / `%s`\n", item.LocalIPCTransport, item.DataRoot)
	fmt.Fprintf(b, "- Update mechanism: `%s`\n", item.UpdateMechanism)
}
