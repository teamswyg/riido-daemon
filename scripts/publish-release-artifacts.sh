#!/usr/bin/env bash
set -euo pipefail

mode="${1:?mode is required}"

case "$mode" in
  checksum)
    cd dist
    sha256sum riido-daemon_* > SHA256SUMS
    ;;
  release)
    notes_file="$RUNNER_TEMP/riido-daemon-release-notes.md"
    cat > "$notes_file" <<EOF
riido-daemon ${GITHUB_REF_NAME} binary release.

Install on macOS/Linux:

\`\`\`bash
curl -fsSL https://raw.githubusercontent.com/teamswyg/riido-daemon/main/scripts/install-riido-daemon.sh | sh
\`\`\`

Desktop/MSIX consumers should download the matching GitHub Release asset into
app user data and launch it without requiring administrator privileges.
EOF
    gh release create "$GITHUB_REF_NAME" dist/* \
      --repo "$GITHUB_REPOSITORY" \
      --title "riido-daemon $GITHUB_REF_NAME" \
      --notes-file "$notes_file" \
      --prerelease
    ;;
  *)
    echo "unsupported publish mode: $mode" >&2
    exit 1
    ;;
esac
