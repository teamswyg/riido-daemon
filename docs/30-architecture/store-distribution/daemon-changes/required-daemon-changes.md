# Required Daemon Changes

[Back to Daemon Changes](../daemon-changes.md)

| Work | Owner context | Output |
| --- | --- | --- |
| Local IPC abstraction | C11 | Unix socket / Windows named pipe adapters behind one port |
| App data root abstraction | C11/C6 | dev-local / Developer ID / App Store / MSIX path selection |
| ExternalToolRegistry | C11/C3 | provider path provenance + version/login status |
| ConsentLedger | C11/C7 | background/provider/workspace/telemetry grants |
| WorkspaceGrantStore | C11/C6 | macOS security-scoped bookmark / Windows folder grant representation |
| StoreChannelPolicyGate | C11/C7 | channel-specific blocked reasons |
| Store-safe demo mode | C11 | `EvaluateReviewDemoMode` + local API `review-demo` reviewable/offline UX without provider spawn or SaaS telemetry sync |
| Distribution contract gate | C11 | executable check that provider binaries are not bundled |
