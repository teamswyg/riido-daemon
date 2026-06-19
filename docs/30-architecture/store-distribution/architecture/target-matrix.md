# Store Distribution Architecture: Target Matrix

[Back to architecture](../architecture.md)

| Target | Artifact | Status | Core blocker |
| --- | --- | --- | --- |
| `developer-id` | signed/notarized macOS app + helper | preferred first | signing/notarization pipeline, helper consent UI |
| `mac-app-store` | sandboxed Mac App Store app | requires redesign | App Sandbox, app group/helper, security-scoped workspace grant, no direct LaunchAgent install |
| `msix-sideload` | signed MSIX | preferred first | Windows named pipe, package local data, manifest/signing |
| `msix-store` | Microsoft Store MSIX packaged desktop app | requires policy gate | runFullTrust explanation, no service install by default, privacy/review notes |
| `dev-local` | `go run` / launchd plist | existing | not a store artifact |

The helper binary that Desktop/MSIX launchers download is published as a GitHub
Release asset by [`release-artifacts.md`](../../release-artifacts.md).

Store package updates remain a Desktop packaging concern. The daemon release
asset is the user-data helper binary source and must not contain provider CLIs
or secrets.
