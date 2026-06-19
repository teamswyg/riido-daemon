# Supervisor Migration Scope

[Back to Integration Gates](../integration-gates.md)

RIID-4662 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/agentbridge/supervisor` 다.

이 package 는 Daemon tier control loop 로서 RuntimeActor pool registration / heartbeat,
task claim, pre-submit C5 eligibility, workdir preparation, EventIngestor append
delegation, terminal result reporting, and shutdown cancellation/archive 를 연결한다.

RIID-4662 당시에는 `controlplane/saasplane`, `controlplane/taskdbplane`,
task DB/project/mwsd/local API, server HTTP transport, infra/secret/state files 를 후속
migration slice 또는 private repo 가 맡기로 남겼다.
