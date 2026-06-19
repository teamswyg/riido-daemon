# RIID-4890: Detect-Selected Executable Start Parity

[Back to figma-boundary-provenance.md](../figma-boundary-provenance.md)

This slice closes the provider runtime gap where capability detection could
select one executable path but process start could re-resolve a different
same-name binary from `PATH`.

This slice:

- adds provider-neutral `StartRequest.Executable`
- passes the selected executable from `bridge.Run` and `runtimeactor.Submit` into provider `BuildStart`
- makes Claude, Codex, OpenClaw, and Cursor command builders prefer `StartOptions.Executable`, then `StartRequest.Executable`, then provider default executable name
- updates OpenClaw integration coverage so the real prompt roundtrip starts the executable that passed the calendar-version detect gate
- preserves the env override rule: explicit `RIIDO_<PROVIDER>_PATH` remains a pin, not a hint

It does not install provider CLIs, change provider auth, add SaaS endpoints,
change assignment polling, or make daemon responsible for provider binary
distribution.
