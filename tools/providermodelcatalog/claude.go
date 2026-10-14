package main

import "strings"

func claudeModels() ([]model, error) {
	body, err := commandOutput("claude", "--help")
	if err != nil {
		return nil, err
	}
	return parseClaudeModels(string(body)), nil
}

func parseClaudeModels(body string) []model {
	rows := make([]model, 0)
	for token := range strings.SplitSeq(claudeModelSection(body), "'") {
		row, ok := claudeModelAlias(token)
		if ok {
			rows = append(rows, row)
		}
	}
	return rows
}

func claudeModelSection(body string) string {
	start := strings.Index(body, "--model <model>")
	if start < 0 {
		return ""
	}
	section := body[start:]
	if before, _, ok := strings.Cut(section, "\n  -n, --name"); ok {
		return before
	}
	return section
}

func claudeModelAlias(alias string) (model, bool) {
	alias = strings.TrimSpace(alias)
	if alias != "fable" && alias != "opus" && alias != "sonnet" &&
		!strings.HasPrefix(alias, "claude-") {
		return model{}, false
	}
	return model{ModelID: alias, Label: modelLabel(alias)}, true
}
