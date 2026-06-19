package main

import "strings"

func decisionsDoc(m manifest) string {
	return renderDoc("Store Distribution Architecture: Decisions", "architecture", func(b *strings.Builder) {
		b.WriteString("Non-goals:\n\n")
		writeBullets(b, []string{
			"C11 domain decisions remain in docs/20-domain/distribution-host-integration.md.",
			"Security policy remains in docs/20-domain/security.md.",
			"Provider capability shared contracts remain in public riido-contracts.",
			"SaaS control-plane and review account details remain in public riido-control-plane.",
		})
		b.WriteString("\nDecisions:\n\n")
		for i, item := range m.Decisions {
			b.WriteString(numbered(i, item))
		}
	})
}

func numbered(i int, item string) string {
	return string(rune('1'+i)) + ". " + item + "\n"
}
