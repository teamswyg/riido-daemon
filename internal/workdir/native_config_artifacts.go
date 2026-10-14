package workdir

import "io/fs"

type nativeConfigArtifact struct {
	Path    string
	Content []byte
	Mode    fs.FileMode
}

func renderProviderNativeConfigArtifacts(plan ProviderNativeConfigPlan) []nativeConfigArtifact {
	var artifacts []nativeConfigArtifact
	artifacts = append(artifacts, providerSettingsArtifacts(plan)...)
	artifacts = append(artifacts, providerHookArtifacts(plan)...)
	return artifacts
}

func providerSettingsArtifacts(plan ProviderNativeConfigPlan) []nativeConfigArtifact {
	var artifacts []nativeConfigArtifact
	for _, path := range plan.ProviderSettingsFiles {
		switch path {
		case ".claude/settings.json":
			artifacts = append(artifacts, nativeConfigArtifact{
				Path:    path,
				Content: []byte(claudeSettingsJSON()),
				Mode:    0o644,
			})
		case ".codex/config.toml":
			artifacts = append(artifacts, nativeConfigArtifact{
				Path:    path,
				Content: []byte(codexConfigTOML()),
				Mode:    0o644,
			})
		}
	}
	return artifacts
}

func providerHookArtifacts(plan ProviderNativeConfigPlan) []nativeConfigArtifact {
	var artifacts []nativeConfigArtifact
	for _, path := range plan.HookFiles {
		if path == ".riido/hooks/claude-audit-hook.sh" {
			artifacts = append(artifacts, nativeConfigArtifact{
				Path:    path,
				Content: []byte(claudeAuditHookScript()),
				Mode:    0o755,
			})
		}
	}
	return artifacts
}
