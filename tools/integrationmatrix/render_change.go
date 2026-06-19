package main

import "strings"

func renderChangeProcedure(m manifest) string {
	return renderDoc("Change Procedure", "../integration-matrix.md", func(b *strings.Builder) {
		writeBullets(b, m.ChangeProcedure)
		b.WriteString("Source manifests:\n\n")
		writeBullets(b, []string{m.ProviderValidationManifest, m.RealCLIObservationManifest})
	})
}
