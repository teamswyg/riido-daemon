# Release Artifacts

> This document owns how the public `riido-daemon` binary is packaged for
> GitHub Releases and consumed by Desktop/MSIX launchers.

## Decision

`riido-daemon` publishes OS/architecture-specific GitHub Release assets from
`v*` tags. The release asset is the public distribution artifact for the daemon
CLI/helper binary. Provider CLIs remain external user-installed tools and are
never bundled in the release archive.

The release workflow is `.github/workflows/release-artifacts.yml`. It runs a
packaging dry-run on pull requests and branch pushes that touch release-owned
paths. It publishes GitHub Release assets only from `v*` tags.

## Asset Names

Release assets are stable within each release tag:

| Platform | Asset |
| --- | --- |
| macOS amd64 | `riido-daemon_darwin_amd64.tar.gz` |
| macOS arm64 | `riido-daemon_darwin_arm64.tar.gz` |
| Linux amd64 | `riido-daemon_linux_amd64.tar.gz` |
| Linux arm64 | `riido-daemon_linux_arm64.tar.gz` |
| Windows amd64 | `riido-daemon_windows_amd64.zip` |
| Windows arm64 | `riido-daemon_windows_arm64.zip` |
| Checksums | `SHA256SUMS` |

Archives contain:

- `riido` or `riido.exe`
- `LICENSE`
- `NOTICE.md`
- `VERSION`

Windows assets use the same public release channel so MSIX/Desktop launchers can
select an artifact by platform. The current Windows process adapter is a
stdlib-only portability layer: it can launch and stop the direct child process,
while full Windows process-tree control remains a separate native-hosting
improvement.

## Installer

macOS/Linux users can install the latest release with:

```bash
curl -fsSL https://raw.githubusercontent.com/teamswyg/riido-daemon/main/scripts/install-riido-daemon.sh | sh
```

The script:

1. detects `darwin`/`linux` and `amd64`/`arm64`;
2. downloads the matching GitHub Release asset and `SHA256SUMS`;
3. verifies the checksum;
4. installs `riido` to `$HOME/.riido/bin` unless `RIIDO_DAEMON_INSTALL_DIR` is
   set.

Use a specific version with:

```bash
RIIDO_DAEMON_VERSION=v0.0.1 \
curl -fsSL https://raw.githubusercontent.com/teamswyg/riido-daemon/main/scripts/install-riido-daemon.sh | sh
```

## Desktop And MSIX Consumption

Riido Desktop should treat GitHub Release assets like an open-source helper
binary source:

1. select the asset by platform and architecture;
2. download it over HTTPS into the app user data area;
3. verify `SHA256SUMS`;
4. extract the binary under user-writable app data;
5. launch the daemon with `RIIDO_DEVICE_ID`, `RIIDO_DEVICE_SECRET`, and
   `RIIDO_SAAS_URL`.

The Desktop/MSIX launcher must not install the daemon under Program Files or
another administrator-owned path. Microsoft Store/MSIX packaging should keep the
downloaded daemon under package local data or another user-writable app data
root, then execute it as the current user.

This release channel does not replace Store package update rules. Store app
updates remain owned by the Desktop packaging target. The daemon release asset
is only the helper binary that the launcher chooses and runs.

### CDN latest mirror

GitHub Release assets are the immutable daemon binary source. The CDN path
`https://cdn.riido.io/releases/latest/ai-agent/` is a mutable Desktop
development/test mirror of those release assets, not a separate build source.

When Desktop consumes the CDN mirror, the operator must update it from a tagged
GitHub Release asset with the same archive name, then invalidate the CloudFront
paths before asking client developers to retest:

- `releases/latest/ai-agent/riido-daemon_darwin_arm64.tar.gz`
- `releases/latest/ai-agent/riido-daemon_darwin_amd64.tar.gz`

The archive `VERSION` file must identify the release tag that was mirrored. A
stale CDN mirror is a release defect even when GitHub Releases already contain a
newer daemon tag.

## Review Boundary

The release archive must not include:

- Claude, Codex, OpenClaw, Cursor, or any other provider CLI binary;
- provider tokens, API keys, or environment files;
- workspace files or user data;
- signing credentials or deployment evidence.

These constraints are inherited from
[`distribution-host-integration.md`](../20-domain/distribution-host-integration.md)
and [`store-distribution.md`](store-distribution.md).
