# README: Module Map

[Back to README](../../README.md)

- `cmd/riido`: CLI/local daemon adapter that parses flag/env and assembles local
  surfaces.
- `internal/agentbridge`: provider-neutral run/request/event/result domain.
- `internal/agentbridge/session`: one-run session actor.
- `internal/agentbridge/runtimeactor`: runtime mailbox, capability
  reconciliation, slot/heartbeat/stop boundary.
- `internal/agentbridge/supervisor`: task claim, runtime dispatch, workdir
  preparation, event ingest, result reporting loop.
- `internal/agentbridge/controlplane/saasplane`: SaaS assignment polling and
  reporting adapter.
- `internal/agentbridge/controlplane/taskdbplane`: local task DB source/reporter.
- `internal/provider/{claude,codex,openclaw,cursor}`: external CLI adapters.
- `internal/hostintegration`: Store/host integration pure model.
- `internal/riidoapi`: local IPC API over Unix socket and Windows named pipe.
- `internal/taskdb`: public daemon copy of `riido-task-db.v1` guarded mutation
  adapter.
- `internal/mwsdbridge`, `internal/project`: macmini-workspace bridge and
  workspace/task projection.
- `packaging/store`, `tools/storecontract`: Store distribution executable
  contract and verifier.
