package main

import (
	"fmt"
	"strings"
)

func renderUpstreamProvenance(m manifest) string {
	return renderDoc("Upstream Provenance", func(b *strings.Builder) {
		p := m.SourceCoverageManifestProvenance
		fmt.Fprintf(b, "`source_coverage_manifest_provenance.stabilized_by` mirrors upstream coverage history from %s: `%s`.\n\n", p.Repo, strings.Join(p.StabilizedBy, "`, `"))
		fmt.Fprintf(b, "`%s` made `%s` a source field. Daemon records `mirrors_source_field = %q` and `source_field_introduced_by = %q` so local projection does not redefine upstream history.\n\n", p.SourceFieldIntroducedBy, p.MirrorsSourceField, p.MirrorsSourceField, p.SourceFieldIntroducedBy)
		renderLimitations(b, m.MirroredSupportingToolLimitations)
	})
}

func renderLimitations(b *strings.Builder, limitations []toolLimitation) {
	for _, limitation := range limitations {
		fmt.Fprintf(b, "The `%s` limitation is local mirror evidence for %s. It comes from %s stabilized by `%s`.\n\n",
			limitation.SourceID,
			limitation.LocalRiidoTask,
			limitation.SourceOwner,
			strings.Join(limitation.SourceStabilizedBy, "`, `"),
		)
		fmt.Fprintf(b, "Required authoritative pages: `%s`.\n\n", strings.Join(limitation.RequiredAuthoritativePages, "`, `"))
		fmt.Fprintf(b, "Non-UI nodes preserved by this limitation: `%s`.\n\n", strings.Join(limitation.MustPreserveNonUINodes, "`, `"))
	}
}
