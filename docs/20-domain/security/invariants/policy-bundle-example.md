# Security Invariants: Policy Bundle Example

[Back to invariants](../invariants.md)

This example records the current `riido-policy-bundle.v1` JSON shape.

```json
{
  "schema_version": "riido-policy-bundle.v1",
  "version": "policy-bundle.example.v1",
  "effective_since": "2026-05-27T00:00:00Z",
  "superseded_at": null,
  "trust_tier_policies": {
    "IsolatedContainer": {
      "allowed_surfaces": {
        "unsafe_bypass": [
          "codex:--yolo"
        ],
        "native_config_hooks": [
          "claude:command-hooks:audit"
        ],
        "native_config_files": [],
        "tool_use": [
          "tool:network-egress"
        ]
      }
    }
  }
}
```
