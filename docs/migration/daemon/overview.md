# Riido Daemon Migration Plan: Overview

[Back to daemon.md](../daemon.md)


> Riido task: RIID-4636 `[Daemon] 기존 riido_daemon daemon 마이그레이션 계획/문서화`

This document defines how the daemon/runtime part of the former private
`riido_daemon` repository moves into the public `riido-daemon` repository.

## Goal

`riido-daemon` owns the customer-PC daemon runtime, local host integration, and
provider execution boundary. It must stay public, store-reviewable, and
free of non-Riido dependencies unless a later ADR explicitly changes that rule.

## Retired Historical Source Boundary

The first migration wave originally compared against the former private
`riido_daemon` source. That source is now **retired history only**.

New work must not read, compare, copy from, cherry-pick from, push to, open PRs
against, merge, or otherwise modify `riido_daemon_private` /
`riido-daemon-private`. If a required fact is missing from public
`riido-daemon`, the fact must be written into this public SSOT or promoted to
`riido-contracts`; the private repository is not a fallback source of truth.

The historical daemon slice was the code and documentation that implement these
bounded contexts:

| Context | Source paths |
| --- | --- |
| C3 Provider Capability | `internal/provider/capability` |
| C4 Provider Runtime / Adapter | `internal/agentbridge`, `internal/provider/{claude,codex,openclaw,cursor}`, `pkg/agentbridge`, `internal/process` |
| C5 Runtime Scheduling | `internal/scheduling`, `internal/agentbridge/runtimeactor`, `internal/agentbridge/supervisor` |
| C6 Workspace / Native Config | `internal/workdir` |
| C7 Security / Policy | `internal/policy` |
| C8 Validation | `internal/validation` |
| C9 Local Locking | `internal/lock` |
| Local workspace projection / mwsd ACL | `internal/mwsdbridge`, `internal/project` |
| C11 Distribution / Host Integration | `internal/hostintegration`, local transport pieces in `internal/riidoapi`, `packaging/store`, `tools/storecontract`, `NOTICE.md` |

The documentation source is:

- `docs/20-domain/provider-runtime.md`
- `docs/20-domain/provider-capability.md`
- `docs/20-domain/security.md`
- `docs/20-domain/security-redaction.md`
- `docs/20-domain/distribution-host-integration.md`
- `docs/30-architecture/module-decomposition.md`
- `docs/30-architecture/integration-matrix.md`
- `docs/30-architecture/config-reference.md`
- `docs/30-architecture/store-distribution.md`
- daemon-related roadmap/audit files under `docs/50-roadmap/`

## Target Boundary

Move into `riido-daemon`:

- provider-neutral runtime actors and session actors
- provider adapter ACLs for Claude, Codex, OpenClaw, and Cursor
- process spawning ports and fakes
- local-only daemon control surfaces
- host integration models for store-safe local execution
- daemon-side validation and black-box tests
- daemon SSOT docs and daemon-specific ADRs

Do not move into `riido-daemon`:

- `cmd/riido_ai_server` or `internal/riidoaiserver`
- Terraform, AWS, ECS, ECR, WAF, ACM, Route53, or release evidence workflows
- `.riido-local`, state files, credentials, account IDs, or deploy artifacts
- shared contract code that must be consumed by both daemon and control-plane
- bundled Claude/Codex/OpenClaw/Cursor CLI binaries

## Migration Order

1. Port SSOT docs first.
   Keep domain decisions in docs before moving code that executes them.

2. Move provider-neutral primitives.
   Start with `internal/agentbridge` root types, reducer, command/result/event
   contracts, and tests that do not import concrete providers.

3. Move process/workdir/policy/validation support packages.
   Keep adapters behind ports and preserve stdlib-only verification.

4. Move provider adapters one at a time.
   Each provider migration must include parser/golden tests and detect command
   tests. Real CLI integration tests remain opt-in and skipped unless the CLI is
   installed.

5. Move daemon runtime actors and local host integration.
   Supervisor/runtime/session actors should remain mailbox-owned. Do not add
   shared mutable state as a migration shortcut.

6. Rebuild daemon workflows in the public repo.
   Public CI should run unit, domain, generated-drift, dependency, and
   black-box daemon checks. Private CI should not duplicate those expensive
   checks.

## Current Migration Slices

### RIID-4643 — contracts import gate

`riido-daemon` started by consuming
`github.com/teamswyg/riido-contracts v0.1.0` and keeping CI limited to
Riido-owned Go module dependencies. This was a compatibility gate only; it did
not move runtime packages. Later slices may bump the module version when a
newly migrated daemon adapter needs a newer shared contract.

### RIID-4645 — local process / validation / lock core

This slice moves provider-neutral local daemon primitives that have no external
dependencies:

- `internal/process` and `internal/process/processexec`
- `internal/validation`
- `internal/lock`
- `internal/logging`
- `internal/jsontest`
- C8 validation and C9 locking SSOT docs under `docs/20-domain/`

This slice does not move provider adapters, runtime/session/supervisor actors,
task DB/project/mwsd/local API packages, CLI commands, private infra, secrets,
or local machine state.

### RIID-4646 — runtime scheduling domain

This slice moves the pure C5 scheduling domain:

- `internal/scheduling`
- `docs/20-domain/runtime-scheduling.md`

The package imports provider capability types from
`github.com/teamswyg/riido-contracts/provider/capability`; the module version
is the current `go.mod` contract version.

This slice does not move supervisor/runtimeactor/session/provider adapters,
task DB/project/mwsd/local API packages, provider process execution, private
infra, secrets, or local machine state.

### RIID-4647 — workspace native config domain

This slice moves the pure C6 workspace / native config domain:

- `internal/workdir`
- `docs/20-domain/workspace.md`
- native config plan generator support needed by `go generate ./internal/workdir`

The package imports IR types from `github.com/teamswyg/riido-contracts/ir`; the
module version is the current `go.mod` contract version.

This slice does not move provider adapters, runtime/session/supervisor actors,
C7 policy/security implementation, C11 host integration implementation,
task DB/project/mwsd/local API packages, private infra, secrets, or local
machine state.

### RIID-4964 — active assignment resume recovery

This slice closes the daemon side of the provider-session recovery path for
locally lost in-flight state. When the control plane returns an active
assignment that already has a pinned `provider_session_id`, the daemon now uses
that provider session as `TaskRequest.ResumeSessionID` before falling back to
the original `resume_session_id`.

This slice does not add a durable run-attempt table, recovery mode enum, or web
approval handoff. Those remain follow-up lifecycle/FSM work.
