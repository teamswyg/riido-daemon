package main

import "strings"

func renderVerification(m manifest) string {
	return renderDoc("Verification", func(b *strings.Builder) {
		b.WriteString("`go test ./tools/figmaboundary -count=1` verifies:\n\n")
		writeBullets(b, verificationBullets(m))
		b.WriteString("`go run ./tools/figmaboundarydocs -check-doc` verifies that these reader docs are generated from the boundary manifest and entry catalog.\n")
	})
}

func verificationBullets(m manifest) []string {
	return []string{
		"manifest schema, RIID, Figma file/page identity",
		"full upstream coverage provenance and source-field marker",
		"metadata page-list limitation provenance",
		"authoritative pages " + sentenceList(requiredPages(m)),
		"non-UI daemon evidence nodes are preserved",
		"all daemon-relevant nodes remain in entry files",
		"every entry separates `daemon_scope`, upstream owners, daemon consumed facts, and client-owned facts",
		"stale agent settings node/template wording does not return",
		"context/provider-runtime/daemon/CLI docs link this boundary",
	}
}

func requiredPages(m manifest) []string {
	if len(m.MirroredSupportingToolLimitations) == 0 {
		return nil
	}
	return m.MirroredSupportingToolLimitations[0].RequiredAuthoritativePages
}
