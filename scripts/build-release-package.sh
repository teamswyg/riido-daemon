#!/usr/bin/env bash
set -euo pipefail

target="${1:?target is required}"
IFS=/ read -r goos goarch format <<< "$target"
version="${GITHUB_REF_NAME:-pull-request}"
out_dir="dist/${goos}_${goarch}"
package_dir="$RUNNER_TEMP/package-${goos}-${goarch}"

mkdir -p "$out_dir" "$package_dir"

binary="riido"
if [ "$goos" = "windows" ]; then
  binary="riido.exe"
fi

GOOS="$goos" GOARCH="$goarch" go build \
  -trimpath -ldflags "-s -w -X main.binaryVersion=$version" -o "$package_dir/$binary" ./cmd/riido
cp LICENSE NOTICE.md "$package_dir/"
printf '%s\n' "$version" > "$package_dir/VERSION"

case "$format" in
  zip)
    asset="riido-daemon_${goos}_${goarch}.zip"
    (cd "$package_dir" && zip -qr "$GITHUB_WORKSPACE/$out_dir/$asset" .)
    ;;
  tar)
    asset="riido-daemon_${goos}_${goarch}.tar.gz"
    tar -C "$package_dir" -czf "$out_dir/$asset" .
    ;;
  *)
    echo "unsupported package format: $format" >&2
    exit 1
    ;;
esac

{
  echo "RIIDO_RELEASE_ARTIFACT_NAME=riido-daemon-${goos}-${goarch}"
  echo "RIIDO_RELEASE_ARTIFACT_PATH=$out_dir/*"
} >> "$GITHUB_ENV"
