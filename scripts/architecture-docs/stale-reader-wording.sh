#!/usr/bin/env bash
set -euo pipefail

stale_pattern='후속 architecture-doc migration|architecture-doc migration slice|github.com/teamswyg/riido_daemon|cmd/riido_ai_server|internal/riidoaiserver|terraform/'
reader_list="$(mktemp)"
stale=""
trap 'rm -f "$reader_list"' EXIT

find docs/20-domain docs/30-architecture docs/50-roadmap -type f -name '*.md' >"$reader_list"

if [ -s "$reader_list" ]; then
  stale="$(while IFS= read -r path; do grep -H --line-number -E "$stale_pattern" "$path" || true; done <"$reader_list")"
fi

if [ -n "$stale" ]; then
  printf 'Stale or cross-boundary architecture reader wording:\n%s\n' "$stale"
  exit 1
fi
