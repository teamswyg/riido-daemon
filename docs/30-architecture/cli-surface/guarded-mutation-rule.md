# Guarded Mutation Rule

[Back to CLI Surface SSOT](../cli-surface.md)

`riido task transition`, `riido task evidence`, `riido task validate`, and their
`riido api ...` equivalents must go through the same guarded mutation path used
by the local API. Approval IDs, command IDs, idempotent receipts, replay
mismatch checks, and deterministic validation evidence remain adapter-invariant.
