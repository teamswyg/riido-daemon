#!/usr/bin/env bash
# Download only signed inputs; always rebuild the daemon from the checked-out tag.
set -euo pipefail
arch="${1:?Go architecture required}"
case "$arch" in arm64|amd64) ;; *) exit 1 ;; esac
[[ "${GITHUB_REF:-}" = refs/tags/v* ]] && [ "$(uname -s)" = Darwin ]
version="${GITHUB_REF_NAME:?}"
revision="$(git rev-parse HEAD)"
work="$(mktemp -d "$RUNNER_TEMP/riido-release-input.XXXXXX")"
work="$(cd "$work" && pwd -P)"
asset="riido-macos-input_darwin_$arch.tar.gz"
gh release download "$version" --repo "$GITHUB_REPOSITORY" --dir "$work" --pattern "$asset"
go run ./tools/clihelperpackage -archive "$work/$asset" -version "$version" -arch "$arch" -require-helper -recheck-signature
# The validated archive allows only regular files/directories with bounded safe paths.
tar -xzf "$work/$asset" -C "$work" ./RiidoCLIAuthHelper.app
bundle="$work/RiidoCLIAuthHelper.app"
[ "$(/usr/bin/plutil -extract RiidoCLISourceRevision raw -o - "$bundle/Contents/Info.plist")" = "$revision" ] || {
  echo 'Signed helper source revision does not match release commit' >&2; exit 1;
}
xcrun stapler validate "$bundle"
/usr/sbin/spctl --assess --type execute "$bundle"
printf 'RIIDO_CLI_AUTH_HELPER_BUNDLE=%s\n' "$bundle" >> "$GITHUB_ENV"
