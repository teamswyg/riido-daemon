# Store Review Invariants

[Back to Riido CLI Migration Plan](../cli.md)

- The CLI must not create public network listeners.
- Provider CLIs are discovered external tools, not bundled payloads.
- Commands that mutate guarded task state must preserve approval-id and receipt
  rules.
- Unsafe provider flags must remain policy-gated.
