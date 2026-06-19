# Decision and Asset Names

[Back to release artifacts](../release-artifacts.md)

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
