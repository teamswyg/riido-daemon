# Change Procedure

[Back to Module Decomposition SSOT](../module-decomposition.md)

When adding a package, env var, CLI flag, or adapter:

1. Update this document if the package/dependency map changes.
2. Update `docs/20-domain/context-map.md` if bounded-context ownership changes.
3. Update `docs/30-architecture/config-reference.md` for env/config changes.
4. Add or update a focused public workflow when the boundary can be checked in
   GitHub Actions without secrets or provider binaries.
