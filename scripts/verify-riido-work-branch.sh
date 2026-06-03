#!/usr/bin/env bash
set -euo pipefail

branch="${1:-}"

if [[ -z "$branch" ]]; then
  echo "usage: $0 <branch-name>" >&2
  exit 2
fi

if [[ "$branch" == "main" ]]; then
  exit 0
fi

if [[ "$branch" == *"/"* ]]; then
  echo "branch must be the exact Riido branchName, not a namespaced local helper branch: $branch" >&2
  exit 1
fi

if [[ "$branch" =~ [[:space:]] ]]; then
  echo "branch must not contain whitespace: $branch" >&2
  exit 1
fi

if [[ ! "$branch" =~ ^[A-Z][A-Z0-9]*-[0-9]+-.+$ ]]; then
  echo "branch must start with a Riido task key and slug, for example A-40-AI-Agent-SSOT-Riido-branchName" >&2
  echo "actual: $branch" >&2
  exit 1
fi

