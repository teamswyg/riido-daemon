package main

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	claudeprovider "github.com/teamswyg/riido-daemon/internal/provider/claude"
)

func claudeRuntimeModels() []runtimeactor.RuntimeModel {
	if models := generatedProviderRuntimeModels(claudeprovider.Name, "sonnet"); len(models) > 0 {
		return models
	}
	body := runtimeModelCommandOutput(claudeprovider.DefaultExecutable, "--help")
	return parseClaudeRuntimeModelHelp(body)
}

func parseClaudeRuntimeModelHelp(body []byte) []runtimeactor.RuntimeModel {
	section := claudeModelHelpSection(string(body))
	models := make([]runtimeactor.RuntimeModel, 0)
	for token := range strings.SplitSeq(section, "'") {
		model, ok := claudeRuntimeModelAlias(token)
		if ok {
			models = append(models, model)
		}
	}
	return normalizeRuntimeModels(models, "sonnet")
}

func claudeModelHelpSection(help string) string {
	start := strings.Index(help, "--model <model>")
	if start < 0 {
		return ""
	}
	section := help[start:]
	if before, _, ok := strings.Cut(section, "\n  -n, --name"); ok {
		return before
	}
	return section
}

func claudeRuntimeModelAlias(alias string) (runtimeactor.RuntimeModel, bool) {
	alias = strings.TrimSpace(alias)
	if alias != "fable" && alias != "opus" && alias != "sonnet" &&
		!strings.HasPrefix(alias, "claude-") {
		return runtimeactor.RuntimeModel{}, false
	}
	return runtimeModelRecord(alias, claudeRuntimeModelLabel(alias), false)
}

func claudeRuntimeModelLabel(alias string) string {
	parts := strings.Split(alias, "-")
	for i, part := range parts {
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}
