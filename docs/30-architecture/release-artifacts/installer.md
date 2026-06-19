# Installer

[Back to release artifacts](../release-artifacts.md)

macOS/Linux users can install the latest release with:

```bash
curl -fsSL https://raw.githubusercontent.com/teamswyg/riido-daemon/main/scripts/install-riido-daemon.sh | sh
```

The script:

1. detects `darwin`/`linux` and `amd64`/`arm64`;
2. resolves `RIIDO_DAEMON_VERSION=latest` through the GitHub Releases API, so
   the newest Riido pre-release tag is installable even when GitHub's stable
   `/releases/latest` endpoint is empty;
3. downloads the matching GitHub Release asset and `SHA256SUMS`;
4. verifies the checksum;
5. installs `riido` to `$HOME/.riido/bin` unless `RIIDO_DAEMON_INSTALL_DIR` is
   set.

Use a specific version with:

```bash
RIIDO_DAEMON_VERSION=v0.0.1 \
curl -fsSL https://raw.githubusercontent.com/teamswyg/riido-daemon/main/scripts/install-riido-daemon.sh | sh
```
