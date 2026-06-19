#!/usr/bin/env bash
set -euo pipefail

allowed="${1:-github.com/teamswyg/riido-contracts}"
disallowed="$(go list -m all | awk -v allowed="$allowed" \
  'NR > 1 && $1 != allowed { print $1 }')"

if [ -n "$disallowed" ]; then
  printf 'Disallowed Go dependencies:\n%s\n' "$disallowed"
  exit 1
fi
