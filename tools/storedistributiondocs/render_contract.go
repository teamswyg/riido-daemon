package main

import (
	"fmt"
	"strings"
)

func executableContractDoc(m manifest, c contract) string {
	return renderDoc("Executable Contract", "Daemon Changes", func(b *strings.Builder) {
		fmt.Fprintf(b, "The local preflight contract lives at [`../../../%s`](../../../%s).\n\n", m.StoreContract, m.StoreContract)
		b.WriteString("Run the executable contract locally with:\n\n")
		b.WriteString("```bash\n")
		fmt.Fprintf(b, "go run ./tools/storecontract -contract %s -repo .\n", m.StoreContract)
		b.WriteString("```\n\n")
		b.WriteString("The contract validates:\n\n")
		writeBullets(b, []string{
			"provider CLI bundling is explicitly forbidden",
			"required SSOT documents exist",
			"required store channels are declared",
			"store artifact roots do not contain files named like provider executables",
			"store artifact roots do not contain hardcoded developer user paths",
		})
		b.WriteString("\nChannel runtime fields:\n\n")
		for _, item := range c.Channels {
			fmt.Fprintf(b, "- `%s`: `%s`, `%s`, `%s`, `%s`, `%s`\n",
				item.ID, item.RuntimeRole, item.BackgroundRule,
				item.LocalIPCTransport, item.DataRoot, item.UpdateMechanism)
		}
	})
}

func externalSourcesDoc(m manifest) string {
	return renderDoc("External Sources", "Daemon Changes", func(b *strings.Builder) {
		b.WriteString("Policy can change, so these are fact sources, not copied rules:\n\n")
		for _, source := range m.ExternalSources {
			fmt.Fprintf(b, "- %s: <%s>\n", source.Label, source.URL)
		}
		b.WriteString("\nIf these external policies change, this document and the C11 domain SSOT must be updated in the same work unit.\n")
	})
}
