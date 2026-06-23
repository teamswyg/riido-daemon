package main

import (
	"slices"
	"strings"
)

func normalizeModels(rows []model) []model {
	out := make([]model, 0, len(rows))
	seen := make(map[string]bool, len(rows))
	for _, row := range rows {
		modelID := strings.TrimSpace(row.ModelID)
		if modelID == "" || seen[modelID] {
			continue
		}
		seen[modelID] = true
		label := strings.TrimSpace(row.Label)
		if label == "" {
			label = modelID
		}
		out = append(out, model{ModelID: modelID, Label: label})
	}
	slices.SortFunc(out, func(a, b model) int {
		return strings.Compare(a.ModelID, b.ModelID)
	})
	return out
}
