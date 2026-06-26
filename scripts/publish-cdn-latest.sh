#!/usr/bin/env bash
set -euo pipefail

mode="${1:-verify}"
evidence_out=""
[ "${2:-}" = "-evidence-out" ] && evidence_out="${3:?evidence path is required}"
dist="${RIIDO_CDN_DIST_DIR:-dist}"
bucket="${RIIDO_CDN_BUCKET:-riido-production}"
prefix="${RIIDO_CDN_PREFIX:-releases/latest/ai-agent}"
distribution="${RIIDO_CLOUDFRONT_DISTRIBUTION_ID:-EWFLNFS69DNTL}"
dry_run="${RIIDO_CDN_DRY_RUN:-false}"
assets="riido-daemon_darwin_arm64.tar.gz riido-daemon_darwin_amd64.tar.gz SHA256SUMS"

fail() { echo "$*" >&2; exit 1; }

sha256_value() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
    return
  fi
  shasum -a 256 "$1" | awk '{print $1}'
}

check_asset() {
  asset="$1"
  path="$dist/$asset"
  [ -f "$path" ] || fail "missing CDN asset: $path"
  [ "$asset" = "SHA256SUMS" ] && return
  expected="$(awk -v a="$asset" '$2 == a {print $1}' "$dist/SHA256SUMS")"
  [ -n "$expected" ] || fail "missing checksum for $asset"
  [ "$(sha256_value "$path")" = "$expected" ] || fail "checksum mismatch for $asset"
  version="$(tar -xOzf "$path" ./VERSION 2>/dev/null || tar -xOzf "$path" VERSION)"
  [ -n "$version" ] || fail "missing VERSION in $asset"
  if [ -n "${RIIDO_RELEASE_TAG:-}" ] && [ "$version" != "$RIIDO_RELEASE_TAG" ]; then
    fail "$asset VERSION=$version does not match RIIDO_RELEASE_TAG=$RIIDO_RELEASE_TAG"
  fi
}

verify_assets() {
  for asset in $assets; do
    check_asset "$asset"
  done
}

publish_assets() {
  for asset in $assets; do
    if [ "$dry_run" = "true" ]; then
      aws s3 cp "$dist/$asset" "s3://$bucket/$prefix/$asset" \
        --cache-control no-cache --dryrun
      continue
    fi
    aws s3 cp "$dist/$asset" "s3://$bucket/$prefix/$asset" \
      --cache-control no-cache
  done
  if [ "$dry_run" != "true" ] && [ -n "$distribution" ]; then
    aws cloudfront create-invalidation \
      --distribution-id "$distribution" \
      --paths "/$prefix/*"
  fi
}

write_evidence() {
  [ -z "$evidence_out" ] && return
  mkdir -p "$(dirname "$evidence_out")"
  printf '{"schema_version":"riido-cdn-latest-ai-agent.v1","status":"verified","mode":"%s","release_tag":"%s","bucket":"%s","prefix":"%s","dry_run":"%s"}\n' \
    "$mode" "${RIIDO_RELEASE_TAG:-}" "$bucket" "$prefix" "$dry_run" > "$evidence_out"
}

case "$mode" in
  verify) verify_assets ;;
  sync) verify_assets; publish_assets ;;
  *) fail "unsupported mode: $mode" ;;
esac
write_evidence
