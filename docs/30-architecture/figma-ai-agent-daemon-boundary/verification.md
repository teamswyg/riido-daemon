# Verification

[Back to Figma AI Agent Daemon Boundary](../figma-ai-agent-daemon-boundary.md)

`go test ./tools/figmaboundary -count=1` verifies:

- manifest schema, RIID, Figma file/page identity
- full upstream coverage provenance and source-field marker
- metadata page-list limitation provenance
- authoritative pages `129:5215`, `42:3014`, `0:1`
- non-UI daemon evidence nodes are preserved
- all daemon-relevant nodes remain in entry files
- every entry separates `daemon_scope`, upstream owners, daemon consumed facts,
  and client-owned facts
- stale agent settings node/template wording does not return
- context/provider-runtime/daemon/CLI docs link this boundary
