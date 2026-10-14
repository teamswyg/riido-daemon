package project

import "github.com/teamswyg/riido-daemon/internal/mwsdbridge"

func FromMwsdSnapshot(snapshot mwsdbridge.Snapshot) (WorkspaceProjection, error) {
	if err := snapshot.Validate(); err != nil {
		return WorkspaceProjection{}, err
	}

	projection := baseWorkspaceProjection(snapshot)
	appendProjectsFromMwsd(&projection, snapshot.Projects.Repositories)
	projection.DocumentTaskLinks = documentTaskLinks(snapshot.Graph.Documents, projection)
	appendMwsdDiagnostics(&projection, snapshot)
	appendProjectionInvariantDiagnostics(&projection, snapshot)
	ensureProjectionDiagnostics(&projection)
	return projection, nil
}
