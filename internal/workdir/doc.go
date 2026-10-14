// Package workdir is the C6 Workspace adapter: per-task workdir trees
// and provider-native config file injection (CLAUDE.md / AGENTS.md /
// GEMINI.md / ...).
//
// What this package owns:
//   - Workspace tree layout:
//     <root>/<workspace>/tasks/<task>/runs/<run>/
//     {workdir,output,logs,artifacts,native-config,ir}/ and the
//     .gc_meta.json marker plus archive.json manifest used by lifecycle
//     retention.
//   - The generated provider->native-config file plan registry.
//   - The provider-native config manifest materialization evidence
//     written under .riido/native-config-manifest.json.
//   - The 4-section runtime-config template (Identity / CLI catalog /
//     Hard rules / Workflow) from spec §10 Phase 7.
//   - workspace_id enforcement: empty workspace IDs are rejected
//     (multica.md §6.1 "workspace_id 필수").
//
// What this package does NOT own:
//   - The C7 policy bundle that DECIDES what goes into
//     the rule set. workdir just renders what it is given.
//   - Retention TTL evaluation. The daemon decides when to archive or
//     clean up; workdir provides deterministic filesystem helpers.
package workdir
