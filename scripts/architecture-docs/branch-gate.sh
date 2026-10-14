#!/usr/bin/env bash
set -euo pipefail

scripts/verify-riido-work-branch.sh "A-40-AI-Agent-SSOT-Riido-작업-branchName-사용-강제"

if scripts/verify-riido-work-branch.sh "codex/not-allowed" >/tmp/riido-branch-negative-1.log 2>&1; then
  echo "namespaced helper branch unexpectedly passed"
  exit 1
fi

if scripts/verify-riido-work-branch.sh "feature-branch" >/tmp/riido-branch-negative-2.log 2>&1; then
  echo "non-Riido branch unexpectedly passed"
  exit 1
fi
