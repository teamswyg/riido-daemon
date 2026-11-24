package main

import "github.com/teamswyg/riido-daemon/internal/provider/claude"

func claudeModels() ([]model, error) {
	body, err := commandOutput("claude", "--help")
	if err != nil {
		return nil, err
	}
	return parseClaudeModels(string(body)), nil
}

func parseClaudeModels(body string) []model {
	ids := claude.ModelIDsFromHelp(body)
	rows := make([]model, 0, len(ids))
	for _, id := range ids {
		rows = append(rows, model{ModelID: id, Label: modelLabel(id)})
	}
	return rows
}
