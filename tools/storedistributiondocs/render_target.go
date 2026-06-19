package main

import (
	"fmt"
	"strings"
)

func targetMatrixDoc(c contract) string {
	return renderDoc("Store Distribution Architecture: Target Matrix", "architecture", func(b *strings.Builder) {
		b.WriteString("| Target | Platform | Status | Runtime role | IPC | Update |\n")
		b.WriteString("| --- | --- | --- | --- | --- | --- |\n")
		for _, item := range c.Channels {
			fmt.Fprintf(b, "| `%s` | %s | %s | %s | %s | %s |\n",
				item.ID, item.Platform, item.Status, item.RuntimeRole,
				item.LocalIPCTransport, item.UpdateMechanism)
		}
		b.WriteString("| `dev-local` | local | existing | local helper | local IPC | source/runtime managed |\n\n")
		b.WriteString("Release assets are owned by [`release-artifacts.md`](../../release-artifacts.md).\n")
	})
}
