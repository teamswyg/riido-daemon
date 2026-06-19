# RIID-4573 — Workdir Archive/Retention/Cache/Native Config Closure

[Back to macOS Provider CLI Review](../macos-provider-cli-review.md)

This slice closes the public daemon workdir policy discussion by absorbing
`Q-WS-001` through `Q-WS-006` into the C6/C7/runtime-upgrade SSOT:

- local archive default is same-host `keep-in-place`; external archive backends
  require an explicit future adapter/config
- workdir cleanup is disabled by default and only the opt-in TTL env is active;
  there is no implicit size or task-count cleanup
- shared repo cache prune is operator-triggered maintenance only, guarded by
  the short `repo_cache_update.lock`
- native config overlay means per-task materialization; user-global config
  copy/overlay is not a default behavior
- container/VM workdir handoff belongs to the future C4 runtime launcher /
  platform adapter, while C6 only prepares host-side files and manifests
- dirty workdir native-config reinjection threshold is zero; changes after
  `Preparing`/`Running` use the no-silent-upgrade flow and next-run
  recomputation

The slice adds focused public CI for the workdir policy closure and the
existing workdir cleanup/native-config tests.
