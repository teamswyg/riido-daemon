#!/bin/bash
set -euo pipefail
helper_dir="$(cd "$(dirname "$0")" && pwd)"
output_dir="${1:?usage: build.sh OUTPUT_DIRECTORY [arm64|x86_64]}"
architecture="${2:-$(uname -m)}"
case "$architecture" in arm64|x86_64) ;; *) echo "Unsupported architecture" >&2; exit 2 ;; esac
bundle="$output_dir/RiidoCLIAuthHelper.app"
if [ -e "$bundle" ]; then echo "Output bundle already exists" >&2; exit 2; fi
mkdir -p "$bundle/Contents/MacOS"
cp "$helper_dir/Info.plist" "$bundle/Contents/Info.plist"
xcrun swiftc -warnings-as-errors -O -whole-module-optimization -target "$architecture-apple-macos13.0" \
  "$helper_dir"/Sources/*.swift -o "$bundle/Contents/MacOS/riido-cli-auth-helper"
printf '%s\n' "$bundle"
