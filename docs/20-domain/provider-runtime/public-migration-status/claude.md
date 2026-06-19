# Claude Adapter

[Back to Public Migration Status](../public-migration-status.md)

RIID-4658 moved `internal/provider/claude` into public `riido-daemon`.

The package does not bundle Claude Code CLI. It owns external executable
detection, command construction, stream-json parser, raw event translator, stdin
protocol driver, and provider input approval frame builder.

Real Claude CLI execution is gated by `AGENTBRIDGE_INTEGRATION=1`.

A-51 added an integration gate that expects `ResultCompleted` and an expected
file artifact inside daemon-selected workdir. The gate requires local Claude
Code auth/runtime and explicitly uses `PermissionModeAcceptEdits` to allow
edit/write tools. A skipped gate does not prove filesystem side effects.
