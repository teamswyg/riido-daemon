package main

import (
	"fmt"
	"strings"
)

func macOSHelperDoc(c contract) string {
	return renderDoc("Store Distribution Architecture: macOS Helper / Login Item", "architecture", func(b *strings.Builder) {
		writeRuntimeTable(b, []channel{
			channelByID(c, "developer-id"),
			channelByID(c, "mac-app-store"),
		})
		b.WriteString("\nCommon rules:\n\n")
		writeBullets(b, []string{
			"Provider CLIs are not included in the helper bundle.",
			"Background helper consent is sourced from C11 ConsentLedger.",
			"Channel allowance is checked pre-runtime by C7 EvaluateStoreChannelPolicy.",
			"Mac App Store notes explain helper purpose, consent UX, sandbox use, and provider CLI non-bundling.",
		})
	})
}

func windowsRuntimeDoc(c contract) string {
	return renderDoc("Store Distribution Architecture: Windows MSIX Runtime", "architecture", func(b *strings.Builder) {
		writeRuntimeTable(b, []channel{
			channelByID(c, "msix-sideload"),
			channelByID(c, "msix-store"),
		})
		b.WriteString("\nCommon rules:\n\n")
		writeBullets(b, []string{
			"Provider CLIs are not included in the MSIX package.",
			"Local IPC uses Windows named pipe only.",
			"App data and daemon state live under package local data root.",
			"Workspace access reaches runtime only through user folder grant and C11 consent.",
		})
	})
}

func writeRuntimeTable(b *strings.Builder, rows []channel) {
	b.WriteString("| Channel | Runtime shape | Background rule | IPC/data root | Update |\n")
	b.WriteString("| --- | --- | --- | --- | --- |\n")
	for _, item := range rows {
		fmt.Fprintf(b, "| `%s` | %s | %s | %s / %s | %s |\n",
			item.ID, item.RuntimeRole, item.BackgroundRule,
			item.LocalIPCTransport, item.DataRoot, item.UpdateMechanism)
	}
}
