package project

import (
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
)

func appendMwsdDiagnostics(projection *WorkspaceProjection, snapshot mwsdbridge.Snapshot) {
	projection.Diagnostics = append(projection.Diagnostics, liftDiagnostics("domain", snapshot.Domain.Diagnostics)...)
	projection.Diagnostics = append(projection.Diagnostics, liftDiagnostics("harness", snapshot.Harness.Diagnostics)...)
	projection.Diagnostics = append(projection.Diagnostics, liftDiagnostics("orchestration", snapshot.Orchestration.Diagnostics)...)
	projection.Diagnostics = append(projection.Diagnostics, liftDiagnostics("projects", snapshot.Projects.Diagnostics)...)
}

func appendProjectionInvariantDiagnostics(projection *WorkspaceProjection, snapshot mwsdbridge.Snapshot) {
	appendDecisionGateDiagnostic(projection)
	appendRecommendationDiagnostics(projection)
	if projection.DirectionBias {
		projection.Diagnostics = append(projection.Diagnostics, ProjectionDiagnostic{
			Severity: "warning",
			Code:     "orchestration-direction-biased",
			Message:  "orchestration reports a top-down/bottom-up direction bias",
		})
	}
	if snapshot.Status.SSOTConflictCount > 0 {
		projection.Diagnostics = append(projection.Diagnostics, ProjectionDiagnostic{
			Severity: "error",
			Code:     "ssot-conflicts-present",
			Message:  fmt.Sprintf("mwsd status reports %d SSOT conflicts", snapshot.Status.SSOTConflictCount),
		})
	}
}

func ensureProjectionDiagnostics(projection *WorkspaceProjection) {
	if projection.Diagnostics == nil {
		projection.Diagnostics = []ProjectionDiagnostic{}
	}
}
