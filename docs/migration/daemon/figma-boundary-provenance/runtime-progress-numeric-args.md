# RIID-4917: Runtime Progress Numeric Args

[Back to figma-boundary-provenance.md](../figma-boundary-provenance.md)

This slice closes a live development finding where Codex could emit progress
telemetry with an integer `count` argument for code `1102`. The upstream
`riido-contracts/progressmessage` catalog already defines that argument as an
int, but the daemon projection accepted only string-valued args.

Without this slice, malformed or incomplete progress telemetry could fall back
to raw JSON-shaped copy later in the public thread stream.

This slice:

- accepts primitive JSON progress args
- normalizes them into the string metadata shape used by the SaaS assignment event contract
- preserves rendered Korean progress text for `1102` when `count` is numeric
- updates injected telemetry instruction so providers know that `1102` requires `label`, `count`, and `representative_title`
- keeps the public SSE/client response shape unchanged

It does not add new progress codes, change the append-only progress catalog,
change frontend rendering, or alter provider final-answer content.
