#!/usr/bin/env bash
set -euo pipefail

required_files=(
  docs/20-domain/context-map.md
  docs/30-architecture/module-decomposition.md
  docs/30-architecture/cli-surface.md
  docs/30-architecture/config-reference.md
  docs/30-architecture/integration-matrix.md
  docs/30-architecture/compatibility-gate.md
  docs/30-architecture/runtime-upgrade-flow.md
  docs/30-architecture/loop-engineering.md
  docs/30-architecture/loop-engineering.riido.json
  docs/30-architecture/figma-ai-agent-daemon-boundary.md
  docs/30-architecture/figma-ai-agent-daemon-boundary.riido.json
  docs/30-architecture/agent-execution-unresolved-design/assignment-lifecycle-evidence.riido.json
  docs/30-architecture/riido-work-branch-gate.md
  docs/50-roadmap/open-questions.md
)

for path in "${required_files[@]}"; do
  test -f "$path"
done

scripts/verify-riido-work-branch.sh \
  "A-40-AI-Agent-SSOT-Riido-작업-branchName-사용-강제"

if scripts/verify-riido-work-branch.sh "codex/not-allowed"; then
  echo "namespaced helper branch unexpectedly passed"
  exit 1
fi

if scripts/verify-riido-work-branch.sh "feature-branch"; then
  echo "non-Riido branch unexpectedly passed"
  exit 1
fi

stale_pattern='후속 architecture-doc migration|architecture-doc migration slice|github.com/teamswyg/riido_daemon|cmd/riido_ai_server|internal/riidoaiserver|terraform/'
stale="$(grep -R --line-number -E "$stale_pattern" \
  docs/20-domain docs/30-architecture docs/50-roadmap || true)"
if [ -n "$stale" ]; then
  printf 'Stale or cross-boundary architecture wording:\n%s\n' "$stale"
  exit 1
fi

env_keys=(
  RIIDO_CLAUDE_PATH RIIDO_CODEX_PATH RIIDO_OPENCLAW_PATH RIIDO_CURSOR_PATH
  AGENTBRIDGE_INTEGRATION RIIDO_TASK_QUEUE_DIR RIIDO_TASK_DB_SOURCE_PATH
  RIIDO_SAAS_URL RIIDO_POLICY_BUNDLE_PATH
)

for key in "${env_keys[@]}"; do
  grep -q "$key" docs/30-architecture/config-reference.md
done

scripts/verify-go-dependencies.sh

go test ./tools/figmaboundary -count=1
go test ./tools/providervalidation -count=1
go test ./tools/agentexecutionevidence -count=1
go test ./tools/loopevidence -count=1
go run ./tools/loopevidence -check
go test ./...
