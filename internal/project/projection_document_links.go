package project

import (
	"sort"

	"github.com/teamswyg/riido-daemon/internal/mwsdbridge"
)

func documentTaskLinks(documents []mwsdbridge.Document, projection WorkspaceProjection) []DocumentTaskLink {
	projectID := "macmini-workspace"
	if !hasProject(projection.Projects, projectID) && len(projection.Projects) > 0 {
		projectID = projection.Projects[0].ID
	}
	links := make([]DocumentTaskLink, 0, len(documents))
	for _, document := range documents {
		if document.ID == "" {
			continue
		}
		links = append(links, DocumentTaskLink{
			TaskID:                 "task:" + document.ID,
			DocumentID:             document.ID,
			DocumentPath:           document.Path,
			Title:                  document.Title,
			Status:                 document.Status,
			Owner:                  document.Owner,
			ProjectID:              projectID,
			RecommendedProvider:    projection.RecommendedProvider,
			RecommendedDecisionLLM: projection.RecommendedDecisionLLM,
			RequiresHumanApproval:  projection.DecisionGate == "human-approval-required" || projection.NextAction.RequiresHumanApproval,
			HarnessNextDirection:   projection.HarnessNextDirection,
		})
	}
	sort.Slice(links, func(i, j int) bool {
		return links[i].DocumentID < links[j].DocumentID
	})
	return links
}
