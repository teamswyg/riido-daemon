#!/bin/sh
set -eu

repo="${RIIDO_DAEMON_REPO:-teamswyg/riido-daemon}"
version="${RIIDO_DAEMON_VERSION:-latest}"
channel="${RIIDO_DAEMON_CHANNEL:-}"
cdn_base="${RIIDO_DAEMON_CDN_BASE:-https://cdn.riido.io/releases}"
install_dir="${RIIDO_DAEMON_INSTALL_DIR:-}"
link_dir="${RIIDO_DAEMON_LINK_DIR:-$HOME/.riido-staging/bin}"
# production keeps `riido` in $HOME/.riido/bin. The staging test channel lives
# beside it as `riido-staging` (RIID-6036): package files under
# $HOME/.riido-staging/libexec, a marker so the same binary uses staging
# defaults, and only the riido-staging link on PATH.
command_name="riido"
path_dir=""
channel_marker="INSTALL_CHANNEL"
helper_bundle="RiidoCLIAuthHelper.app"
helper_manifest="CLI_AUTH_HELPER.json"
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
# A channel install reads the same verified assets the desktop app uses from
# the CDN instead of GitHub release tags. The legacy latest/staging CDN paths
# are frozen; the active feeds are safe and staging-safe.
cdn_prefix() {
  case "$1" in
    staging) echo "staging-safe" ;;
    latest) echo "safe" ;;
  esac
}
check_channel() {
  case "$channel" in
    "") return ;;
    staging | latest) ;;
    *) fail "RIIDO_DAEMON_CHANNEL must be staging or latest" ;;
  esac
  [ "$version" = "latest" ] || fail "RIIDO_DAEMON_VERSION cannot be combined with RIIDO_DAEMON_CHANNEL"
  case "$cdn_base" in
    https://*) ;;
    *) fail "RIIDO_DAEMON_CDN_BASE must use https" ;;
  esac
}
# The macOS auth helper ships as a pair: the signed bundle and its manifest.
check_helper_pair() {
  if [ -e "$1/$helper_bundle" ] || [ -e "$1/$helper_manifest" ]; then
    [ -d "$1/$helper_bundle" ] && [ -f "$1/$helper_manifest" ] || fail "package has an incomplete $helper_bundle"
  fi
}
copy_tree() {
  if command -v ditto >/dev/null 2>&1; then
    ditto "$1" "$2"
  else
    cp -R "$1" "$2"
  fi
}
install_package_files() {
  src="$1"
  if [ -f "$src/VERSION" ]; then
    cp "$src/VERSION" "$install_dir/VERSION"
  else
    rm -f "$install_dir/VERSION"
  fi
  rm -rf "$install_dir/$helper_bundle.new"
  if [ -d "$src/$helper_bundle" ]; then
    copy_tree "$src/$helper_bundle" "$install_dir/$helper_bundle.new"
    rm -rf "$install_dir/$helper_bundle"
    mv "$install_dir/$helper_bundle.new" "$install_dir/$helper_bundle"
    cp "$src/$helper_manifest" "$install_dir/$helper_manifest"
    return
  fi
  # A package without the helper must not leave an older helper beside the
  # new binary; the CLI rejects a helper from another version.
  rm -rf "$install_dir/$helper_bundle"
  rm -f "$install_dir/$helper_manifest"
}
profile_file() {
  case "$(basename "${SHELL:-sh}")" in
    zsh) echo "$HOME/.zshrc" ;;
    bash)
      if [ "$(uname -s)" = Darwin ]; then echo "$HOME/.bash_profile"; else echo "$HOME/.bashrc"; fi
      ;;
    fish) echo "$HOME/.config/fish/config.fish" ;;
    *) echo "$HOME/.profile" ;;
  esac
}
print_next_steps() {
  case ":${PATH:-}:" in
    *":$path_dir:"*)
      found="$(command -v "$command_name" 2>/dev/null || true)"
      if [ -n "$found" ] && [ "$found" != "$path_dir/$command_name" ]; then
        echo ""
        echo "주의: 터미널에서 먼저 찾는 ${command_name}는 $found 입니다."
        echo "  PATH에서 $path_dir 를 앞쪽으로 옮기거나 오래된 파일을 지우세요."
      fi
      ;;
    *)
      rc="$(profile_file)"
      echo ""
      echo "터미널에서 ${command_name}를 바로 쓰려면 PATH에 추가하세요:"
      case "$rc" in
        *config.fish) echo "  fish_add_path $path_dir" ;;
        *) echo "  echo 'export PATH=\"$path_dir:\$PATH\"' >> $rc && . $rc" ;;
      esac
      ;;
  esac
  echo ""
  echo "다음 단계:"
  echo "  $command_name doctor       설치 상태 확인"
  if [ -d "$install_dir/$helper_bundle" ]; then
    echo "  $command_name auth login   브라우저로 로그인"
  fi
  echo "  $command_name help         자주 쓰는 명령 보기"
  if [ "$channel" = staging ]; then
    echo ""
    echo "staging 테스트용 CLI입니다. production은 riido, staging은 $command_name 로 따로 실행됩니다."
  fi
}
# Point the staging command at the installed binary; production has no link.
link_staging_command() {
  [ "$channel" = staging ] || return 0
  mkdir -p "$link_dir"
  rm -f "$link_dir/$command_name"
  ln -s "$install_dir/riido" "$link_dir/$command_name"
}
write_channel_marker() {
  if [ "$channel" = staging ]; then
    printf 'staging\n' > "$install_dir/$channel_marker"
  else
    rm -f "$install_dir/$channel_marker"
  fi
}
main() {
  need curl
  need install
  need tar
  check_channel
  if [ "$channel" = staging ]; then
    command_name="riido-staging"
    install_dir="${install_dir:-$HOME/.riido-staging/libexec}"
    path_dir="$link_dir"
  else
    install_dir="${install_dir:-$HOME/.riido/bin}"
    path_dir="$install_dir"
  fi
  os="$(detect_os)"
  arch="$(detect_arch)"
  asset="riido-daemon_${os}_${arch}.tar.gz"
  if [ -n "$channel" ]; then
    base_url="$cdn_base/$(cdn_prefix "$channel")/ai-agent"
  else
    resolved_version="$(resolve_version)"
    base_url="https://github.com/$repo/releases/download/$resolved_version"
  fi
  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "$tmp_dir"' EXIT
  curl -fsSL "$base_url/$asset" -o "$tmp_dir/$asset"
  curl -fsSL "$base_url/SHA256SUMS" -o "$tmp_dir/SHA256SUMS"
  sha256_check "$tmp_dir/$asset" "$tmp_dir/SHA256SUMS"
  mkdir -p "$tmp_dir/package"
  tar -xzf "$tmp_dir/$asset" -C "$tmp_dir/package"
  [ -f "$tmp_dir/package/riido" ] || fail "package is missing riido"
  check_helper_pair "$tmp_dir/package"
  if [ -n "$channel" ]; then
    [ -f "$tmp_dir/package/VERSION" ] || fail "channel package is missing VERSION"
    resolved_version="$(head -n 1 "$tmp_dir/package/VERSION")"
  fi
  mkdir -p "$install_dir"
  # The staging link must name the real directory even for a relative override.
  install_dir="$(cd "$install_dir" && pwd)"
  rm -f "$install_dir/riido"
  install -m 0755 "$tmp_dir/package/riido" "$install_dir/riido"
  install_package_files "$tmp_dir/package"
  write_channel_marker
  link_staging_command
  echo "riido-daemon installed: $install_dir/riido"
  [ "$channel" != staging ] || echo "riido-staging command: $link_dir/$command_name"
  echo "riido-daemon version: $resolved_version"
  [ -z "$channel" ] || echo "riido-daemon channel: $channel"
  print_next_steps
}
main "$@"
