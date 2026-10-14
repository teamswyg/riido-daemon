package main

import "strings"

func cursorModels() ([]model, error) {
	body, err := commandOutput("cursor-agent", "models")
	if err != nil {
		return nil, err
	}
	return parseCursorModels(string(body)), nil
}

func parseCursorModels(body string) []model {
	rows := make([]model, 0)
	for line := range strings.SplitSeq(body, "\n") {
		row, ok := parseCursorModelLine(line)
		if ok {
			rows = append(rows, row)
		}
	}
	return rows
}

func parseCursorModelLine(line string) (model, bool) {
	modelID, label, ok := strings.Cut(strings.TrimSpace(line), " - ")
	if !ok {
		return model{}, false
	}
	modelID = strings.TrimSpace(modelID)
	if modelID == "auto" {
		modelID = "cursor-auto"
	}
	label = strings.TrimSpace(strings.TrimSuffix(label, "(current)"))
	return model{ModelID: modelID, Label: label}, modelID != ""
}
