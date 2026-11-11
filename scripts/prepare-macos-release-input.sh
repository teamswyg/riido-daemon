#!/usr/bin/env bash
# Operator signing step. Final daemon binaries are rebuilt by the tag workflow.
set -euo pipefail
version="${1:?release version required}"
arch="${2:?Go architecture required}"
out="${3:?absolute output directory required}"
[[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'Invalid release version' >&2; exit 1; }
case "$arch" in arm64) swift_arch=arm64 ;; amd64) swift_arch=x86_64 ;; *) exit 1 ;; esac
[ "$(uname -s)" = Darwin ] && [[ "$out" = /* ]]
: "${RIIDO_SIGNING_IDENTITY:?existing Developer ID identity required}"
: "${RIIDO_HELPER_PROFILE:?authorized embedded provisioning profile required}"
: "${APPLE_ID:?}" "${APPLE_APP_SPECIFIC_PASSWORD:?}" "${APPLE_TEAM_ID:?}"
[ "$APPLE_TEAM_ID" = 3PVV8DQPL6 ]
repo="$(cd "$(dirname "$0")/.." && pwd)"
cd "$repo"
git diff --quiet HEAD -- || { echo 'Commit release source before signing' >&2; exit 1; }
revision="$(git rev-parse HEAD)"
mkdir -p "$out"
work="$(mktemp -d "${TMPDIR:-/tmp}/riido-signed-input.XXXXXX")"
trap 'rm -rf "$work"' EXIT
RIIDO_CLI_RELEASE_VERSION="$version" RIIDO_CLI_SOURCE_REVISION="$revision" \
  native/cli-auth-helper/build.sh "$work/helper" "$swift_arch"
bundle="$work/helper/RiidoCLIAuthHelper.app"
cp "$RIIDO_HELPER_PROFILE" "$bundle/Contents/embedded.provisionprofile"
/usr/bin/codesign --force --timestamp --options runtime \
  --entitlements native/cli-auth-helper/entitlements.plist \
  --sign "$RIIDO_SIGNING_IDENTITY" "$bundle"
/usr/bin/ditto -c -k --keepParent "$bundle" "$work/helper.zip"
xcrun notarytool submit "$work/helper.zip" --apple-id "$APPLE_ID" \
  --password "$APPLE_APP_SPECIFIC_PASSWORD" --team-id "$APPLE_TEAM_ID" \
  --wait --output-format json > "$out/notarization-$arch.json"
/usr/bin/plutil -extract status raw -o - "$out/notarization-$arch.json" | /usr/bin/grep -qx Accepted
xcrun stapler staple "$bundle"
xcrun stapler validate "$bundle"
/usr/sbin/spctl --assess --type execute "$bundle"
GITHUB_WORKSPACE="$work" RUNNER_TEMP="$work" GITHUB_ENV="$work/env" \
  GITHUB_REF_NAME="$version" GITHUB_REF="refs/tags/$version" GITHUB_EVENT_NAME=push \
  RIIDO_CLI_AUTH_HELPER_BUNDLE="$bundle" scripts/build-release-package.sh "darwin/$arch/tar"
cp "$work/dist/darwin_$arch/riido-daemon_darwin_$arch.tar.gz" "$out/riido-macos-input_darwin_$arch.tar.gz"
printf 'Prepared %s at source %s\n' "$arch" "$revision"
