#!/usr/bin/env bash
set -euo pipefail

env_keys=(
  RIIDO_CLAUDE_PATH
  RIIDO_CODEX_PATH
  RIIDO_OPENCLAW_PATH
  RIIDO_CURSOR_PATH
  AGENTBRIDGE_INTEGRATION
  RIIDO_TASK_QUEUE_DIR
  RIIDO_TASK_DB_SOURCE_PATH
  RIIDO_SAAS_URL
  RIIDO_POLICY_BUNDLE_PATH
)

for key in "${env_keys[@]}"; do
  grep -q "$key" docs/30-architecture/config-reference.md
done
