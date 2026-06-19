#!/bin/sh
set -eu

repo="${RIIDO_DAEMON_REPO:-teamswyg/riido-daemon}"
version="${RIIDO_DAEMON_VERSION:-latest}"
install_dir="${RIIDO_DAEMON_INSTALL_DIR:-$HOME/.riido/bin}"
fail() { printf '%s\n' "$*" >&2; exit 1; }
need() {
  command -v "$1" >/dev/null 2>&1 || fail "$1 is required"
}
detect_os() {
  case "$(uname -s)" in
    Darwin) echo "darwin" ;;
    Linux) echo "linux" ;;
    *) fail "unsupported OS: $(uname -s)" ;;
  esac
}
detect_arch() {
  case "$(uname -m)" in
    x86_64 | amd64) echo "amd64" ;;
    arm64 | aarch64) echo "arm64" ;;
    *) fail "unsupported arch: $(uname -m)" ;;
  esac
}
checksum_value() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
    return
  fi
  need shasum
  shasum -a 256 "$1" | awk '{print $1}'
}
sha256_check() {
  file="$1"
  sums="$2"
  asset_name="$(basename "$file")"
  expected="$(grep "  $asset_name\$" "$sums" | awk '{print $1}')"
  [ -n "$expected" ] || fail "checksum for $asset_name is missing"
  [ "$(checksum_value "$file")" = "$expected" ] || fail "checksum mismatch for $asset_name"
}
resolve_version() {
  if [ "$version" != "latest" ]; then
    echo "$version"
    return
  fi
  latest="$(curl -fsSL -H "Accept: application/vnd.github+json" \
    "https://api.github.com/repos/$repo/releases?per_page=1" |
    sed -n 's/^[[:space:]]*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' |
    head -n 1)"
  [ -n "$latest" ] || fail "could not resolve latest riido-daemon release for $repo"
  echo "$latest"
}
main() {
  need curl
  need install
  need tar
  os="$(detect_os)"
  arch="$(detect_arch)"
  asset="riido-daemon_${os}_${arch}.tar.gz"
  resolved_version="$(resolve_version)"
  base_url="https://github.com/$repo/releases/download/$resolved_version"
  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "$tmp_dir"' EXIT
  curl -fsSL "$base_url/$asset" -o "$tmp_dir/$asset"
  curl -fsSL "$base_url/SHA256SUMS" -o "$tmp_dir/SHA256SUMS"
  sha256_check "$tmp_dir/$asset" "$tmp_dir/SHA256SUMS"
  mkdir -p "$install_dir"
  tar -xzf "$tmp_dir/$asset" -C "$tmp_dir"
  install -m 0755 "$tmp_dir/riido" "$install_dir/riido"
  echo "riido-daemon installed: $install_dir/riido"
  echo "riido-daemon version: $resolved_version"
  echo "Add $install_dir to PATH or launch it from Riido Desktop."
}
main "$@"
