package main

import (
	"fmt"
	"strings"
)

func renderGatePolicy(m manifest) string {
	return renderDoc("Gate Policy", "../integration-matrix.md", func(b *strings.Builder) {
		b.WriteString("Each provider `TestIntegration` is optional until all gates pass:\n\n")
		for i, item := range m.GatePolicy {
			fmt.Fprintf(b, "%d. %s\n", i+1, item)
		}
		b.WriteByte('\n')
		b.WriteString("Global rules from provider validation:\n\n")
		writeBullets(b, m.ProviderValidation.GlobalRules)
		b.WriteString("`PASS` in this matrix means the provider produced the evidence named in `provider-validation-matrix.riido.json`.\n\n")
		b.WriteString("A skipped integration test, a detected binary, or a SaaS completed thread alone is not filesystem side-effect evidence.\n\n")
		fmt.Fprintf(b, "The security decision itself is owned by [`security.md`](%s) §4.3.\n", detailSecurityLink(m))
	})
}

func detailSecurityLink(m manifest) string {
	if m.SecurityDecisionLink == "../20-domain/security.md" {
		return "../../20-domain/security.md"
	}
	return m.SecurityDecisionLink
}
