package main

import "strings"

func packageBoundariesDoc() string {
	return renderDoc("Store Distribution Architecture: Package Boundaries", "architecture", func(b *strings.Builder) {
		b.WriteString("RIID-4570 decision: riido-daemon owns C11 Store App contracts and local helper runtime shape. ")
		b.WriteString("A future desktop/app repository may own the concrete Store App GUI adapter and OS entitlement calls, ")
		b.WriteString("but must consume the C11/local API contracts instead of redefining domain facts.\n\n")
		b.WriteString("| Surface | Owner | Non-owner |\n| --- | --- | --- |\n")
		rows := []string{
			"C11 domain facts and pure models | riido-daemon | Store App GUI repo must not copy/redefine them",
			"Local helper / broker executable | riido-daemon cmd/riido | Store App GUI must not run provider CLIs directly",
			"Local IPC handler and request envelope | riido-daemon | Store App GUI may only be a client",
			"Store distribution executable contract | riido-daemon | Private infra must not weaken public review invariants",
			"Store App native UI and entitlement calls | future desktop/app repository | riido-daemon domain packages do not import GUI frameworks",
			"Signing and submission credentials | private operator/infra environment | public repositories never store secrets",
			"Shared DTO/schema needed by multiple repos | riido-contracts after promotion | no repo may fork the same fact",
		}
		for _, row := range rows {
			b.WriteString("| " + row + " |\n")
		}
	})
}
