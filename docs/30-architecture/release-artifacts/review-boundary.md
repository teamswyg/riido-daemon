# Review Boundary

[Back to release artifacts](../release-artifacts.md)

The release archive must not include:

- Claude, Codex, OpenClaw, Cursor, or any other provider CLI binary;
- provider tokens, API keys, or environment files;
- workspace files or user data;
- signing credentials or deployment evidence.

These constraints are inherited from
[`distribution-host-integration.md`](../../20-domain/distribution-host-integration.md)
and [`store-distribution.md`](../store-distribution.md).
