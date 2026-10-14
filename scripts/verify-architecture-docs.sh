#!/usr/bin/env bash
set -euo pipefail

bash scripts/architecture-docs/required-files.sh
bash scripts/architecture-docs/branch-gate.sh
bash scripts/architecture-docs/stale-reader-wording.sh
bash scripts/architecture-docs/env-keys.sh
scripts/verify-go-dependencies.sh
bash scripts/architecture-docs/tool-checks.sh
go test ./...
