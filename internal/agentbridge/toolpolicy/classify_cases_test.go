package toolpolicy

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func toolClassificationCases() []classificationCase {
	return []classificationCase{
		{"codex shell approval", agentbridge.ToolRef{Kind: "shell"}, policy.ToolUseDestructiveCommand},
		{"claude bash approval", agentbridge.ToolRef{Name: "Bash", Kind: "Bash"}, policy.ToolUseDestructiveCommand},
		{"codex patch apply", agentbridge.ToolRef{Kind: "patch_apply"}, policy.ToolUseProtectedPathWrite},
		{"protected path write", agentbridge.ToolRef{Name: "Write", Args: map[string]string{"path": ".git/config"}}, policy.ToolUseProtectedPathWrite},
		{"network fetch", agentbridge.ToolRef{Name: "WebFetch"}, policy.ToolUseNetworkEgress},
		{"network shell command", agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "curl https://example.com"}}, policy.ToolUseNetworkEgress},
		{"secret token", agentbridge.ToolRef{Name: "Token"}, policy.ToolUseSecretExposure},
		{"secret arg key", agentbridge.ToolRef{Name: "Read", Args: map[string]string{"api_token": "[redacted]"}}, policy.ToolUseSecretExposure},
		{"secret redacted arg value", agentbridge.ToolRef{Name: "Read", Args: map[string]string{"note": "[redacted]"}}, policy.ToolUseSecretExposure},
		{"secret env read shell command", agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "cat .env.local"}}, policy.ToolUseSecretExposure},
		{"secret manager shell command", agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "aws secretsmanager get-secret-value --secret-id prod/api"}}, policy.ToolUseSecretExposure},
		{"destructive shell command", agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "rm -rf .git"}}, policy.ToolUseDestructiveCommand},
		{"protected env write", agentbridge.ToolRef{Name: "Write", Args: map[string]string{"file_path": ".env.production"}}, policy.ToolUseProtectedPathWrite},
		{"protected ssh write", agentbridge.ToolRef{Name: "Write", Args: map[string]string{"path": "~/.ssh/config"}}, policy.ToolUseProtectedPathWrite},
		{"protected env shell write", agentbridge.ToolRef{Kind: "shell", Args: map[string]string{"command": "printf TOKEN=x > .env"}}, policy.ToolUseProtectedPathWrite},
	}
}
