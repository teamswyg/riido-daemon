# Change Procedure

[Back to context-map.md](../context-map.md)

Changing context ownership or dependency direction is a policy-breaking change.
The same PR must update:

- `docs/20-domain/context-map.md`
- the focused `docs/20-domain/context-map/` section that owns the changed fact
- `docs/30-architecture/module-decomposition.md`
- any package/workflow gate that enforces the boundary
