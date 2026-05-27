// Package hostintegration owns the C11 Distribution / Host Integration
// domain model.
//
// It models distribution channels, app data roots, local IPC endpoints,
// consent facts, workspace grants, external provider CLI registration
// provenance, and the privacy boundary for server-facing provider status.
//
// What this package does NOT own:
//   - Provider capability detection details and compatibility semantics beyond
//     reusing C3 value types from riido-contracts/provider/capability.
//   - Provider process execution / session lifecycle -> future C4 adapters.
//   - Filesystem scanning, OS package APIs, security-scoped bookmarks, named
//     pipes, or Unix socket listeners -> future C11 adapters.
//   - Persistence substrate for the registry -> project/store adapters or
//     future host integration storage.
package hostintegration
