# Cursor Adapter

[Back to Public Migration Status](../public-migration-status.md)

RIID-4661 moved `internal/provider/cursor` into public `riido-daemon`.

The package does not bundle Cursor Agent CLI. It owns root-print,
agent-subcommand, legacy-chat launch profile selection, `--yolo` unsafe bypass
policy gate, daemon task workdir headless workspace trust acknowledgement
`--trust`, external executable detection, stream-json parser, and raw event
translator.

Cursor `--trust` prevents an interactive trust prompt for daemon-selected
workdir. It is not the same as tool auto-approval. `--yolo` remains an unsafe
bypass surface and can be used only after C7 policy gate approval.

Real Cursor Agent execution is gated by `AGENTBRIDGE_INTEGRATION=1`.
