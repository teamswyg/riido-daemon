# Store Review Invariants

[Back to Runtime Snapshot Heartbeat](../runtime-snapshot-heartbeat.md)

- Provider CLIs are external tools, not bundled app payloads.
- The daemon must expose local-only IPC, not public TCP listeners.
- Unsafe provider modes are opt-in policy decisions, not defaults.
- Host trust tier must reject unsafe bypass.
- App Store and MSIX helper/runtime contracts stay in C11 docs and tests.
