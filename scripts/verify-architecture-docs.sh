#!/usr/bin/env bash
set -euo pipefail
required_files=(
  docs/20-domain/context-map.md
  docs/README.md
  docs/readme/document-map.md
  docs/readme/document-map.riido.json
  docs/readme/verification.md
  docs/readme/verification.riido.json
  docs/20-domain/provider-runtime/adapter-acl/event-ingestor-contract.md
  docs/20-domain/provider-runtime/adapter-acl/event-ingestor-contract.riido.json
  docs/20-domain/provider-runtime/runtime-responsibility/provider-event-draft.md
  docs/20-domain/provider-runtime/runtime-responsibility/provider-event-draft.riido.json
  docs/30-architecture/module-decomposition.md
  docs/30-architecture/cli-surface.md
  docs/30-architecture/config-reference.md
  docs/30-architecture/integration-matrix.md
  docs/30-architecture/compatibility-gate.md
  docs/30-architecture/runtime-upgrade-flow.md
  docs/30-architecture/loop-engineering.md
  docs/30-architecture/loop-engineering.riido.json
  docs/30-architecture/provider-real-cli-observation.md
  docs/30-architecture/provider-real-cli-observation.riido.json
  docs/30-architecture/runtime-secret-private-evidence.md
  docs/30-architecture/runtime-secret-private-evidence.riido.json
  docs/20-domain/provider-runtime/adapter-draft-fields/idle-watchdog.md
  docs/20-domain/provider-runtime/adapter-draft-fields/idle-watchdog.riido.json
  docs/20-domain/provider-runtime/adapter-draft-fields/run-lifecycle.riido.json docs/20-domain/provider-runtime/adapter-draft-fields/cancel-interrupt-input.riido.json docs/20-domain/provider-runtime/adapter-draft-fields/approval-wait-timeout.riido.json
  docs/30-architecture/figma-ai-agent-daemon-boundary.md
  docs/30-architecture/figma-ai-agent-daemon-boundary.riido.json
  docs/30-architecture/agent-execution-unresolved-design/assignment-lifecycle-evidence.riido.json
  docs/30-architecture/riido-work-branch-gate.md
  docs/50-roadmap/open-questions.md
)
for path in "${required_files[@]}"; do
  test -f "$path"
done
scripts/verify-riido-work-branch.sh "A-40-AI-Agent-SSOT-Riido-작업-branchName-사용-강제"
if scripts/verify-riido-work-branch.sh "codex/not-allowed" >/tmp/riido-branch-negative-1.log 2>&1; then
  echo "namespaced helper branch unexpectedly passed"
  exit 1
fi
if scripts/verify-riido-work-branch.sh "feature-branch" >/tmp/riido-branch-negative-2.log 2>&1; then
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
go test ./tools/figmaboundary ./tools/providervalidation ./tools/agentexecutionevidence ./tools/loopevidence ./tools/redactiondrift ./tools/providerintegrationevidence ./tools/runtimesecretevidence ./tools/docmap ./tools/repoverification ./tools/semanticeventactivity ./tools/eventauthority ./tools/providerdraftmapping ./tools/terminalresultmapping ./tools/shutdownauthority ./tools/approvaltimeout -count=1
go run ./tools/loopevidence -check
go run ./tools/docmap -check
for tool in repoverification semanticeventactivity eventauthority providerdraftmapping terminalresultmapping shutdownauthority approvaltimeout; do
  go run "./tools/$tool" -check-doc
done
go run ./tools/redactiondrift
go run ./tools/providerintegrationevidence -check-doc
go run ./tools/runtimesecretevidence -check-doc
go test ./...
