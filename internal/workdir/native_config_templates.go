package workdir

func claudeSettingsJSON() string {
	return `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PROJECT_DIR}/.riido/hooks/claude-audit-hook.sh",
            "timeout": 30,
            "statusMessage": "Riido audit"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PROJECT_DIR}/.riido/hooks/claude-audit-hook.sh",
            "timeout": 30,
            "statusMessage": "Riido audit"
          }
        ]
      }
    ]
  }
}
`
}

func claudeAuditHookScript() string {
	return `#!/bin/sh
set -eu

project_dir="${CLAUDE_PROJECT_DIR:-$(pwd)}"
event_dir="$project_dir/.riido/hooks"
mkdir -p "$event_dir"
cat >> "$event_dir/claude-hook-events.jsonl"
printf '\n' >> "$event_dir/claude-hook-events.jsonl"
exit 0
`
}

func codexConfigTOML() string {
	return `# Managed by riido-daemon.
# Reserved for future Codex native config materialization.
# Current Codex runs use adapter-owned full-access sandbox selection instead of task-scoped CODEX_HOME.
`
}
