#!/usr/bin/env bash
set -euo pipefail

target="${1:?target is required}"
IFS=/ read -r goos goarch format <<< "$target"
version="${GITHUB_REF_NAME:-pull-request}"
# GitHub pull_request refs use <number>/merge, not a release version.
if [ "${GITHUB_EVENT_NAME:-}" = pull_request ]; then version="pull-request"; fi
helper_bundle="${RIIDO_CLI_AUTH_HELPER_BUNDLE:-}"
case "$target" in
  darwin/amd64/tar|darwin/arm64/tar|linux/amd64/tar|linux/arm64/tar|windows/amd64/zip|windows/arm64/zip) ;;
  *) echo "unsupported release target" >&2; exit 1 ;;
esac
if [[ ! "$version" =~ ^[A-Za-z0-9][A-Za-z0-9._+-]*$ ]] || [ "${#version}" -gt 128 ]; then
  echo "invalid release version" >&2; exit 1
fi
if [ -n "$helper_bundle" ]; then
  if [ "$goos" != darwin ] || [ "$(uname -s)" != Darwin ] || [[ "$helper_bundle" != /* ]]; then
    echo "optional helper requires a macOS host, Darwin target and absolute bundle path" >&2; exit 1
  fi
fi
out_dir="$GITHUB_WORKSPACE/dist/${goos}_${goarch}"
mkdir -p "$out_dir" "$RUNNER_TEMP"
package_dir="$(mktemp -d "$RUNNER_TEMP/riido-package-${goos}-${goarch}.XXXXXX")"
archive_stage="$(mktemp -d "$out_dir/.archive.XXXXXX")"
trap 'rm -rf "$package_dir" "$archive_stage"' EXIT
binary="riido"
[ "$goos" != windows ] || binary="riido.exe"
GOOS="$goos" GOARCH="$goarch" go build \
  -trimpath -ldflags "-s -w -X main.binaryVersion=$version" -o "$package_dir/$binary" ./cmd/riido
cp LICENSE NOTICE.md "$package_dir/"
printf '%s\n' "$version" > "$package_dir/VERSION"
if [ -n "$helper_bundle" ]; then
  go run ./tools/clihelperpackage -bundle "$helper_bundle" -out "$package_dir" -version "$version" -arch "$goarch"
fi
case "$format" in
  zip)
    asset="riido-daemon_${goos}_${goarch}.zip"
    (cd "$package_dir" && zip -qr "$archive_stage/$asset" .)
    ;;
  tar)
    asset="riido-daemon_${goos}_${goarch}.tar.gz"
    # Avoid macOS resource-fork sidecars; signed bundle resources remain intact.
    COPYFILE_DISABLE=1 tar -C "$package_dir" -czf "$archive_stage/$asset" .
    ;;
esac
if [ -n "$helper_bundle" ]; then
  go run ./tools/clihelperpackage -archive "$archive_stage/$asset" -version "$version" -arch "$goarch" -recheck-signature
fi
mv -f "$archive_stage/$asset" "$out_dir/$asset"
{
  echo "RIIDO_RELEASE_ARTIFACT_NAME=riido-daemon-${goos}-${goarch}"
  echo "RIIDO_RELEASE_ARTIFACT_PATH=dist/${goos}_${goarch}/*"
} >> "$GITHUB_ENV"
