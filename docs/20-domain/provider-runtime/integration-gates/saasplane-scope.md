# SaaS Plane Migration Scope

[Back to Integration Gates](../integration-gates.md)

RIID-4689 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/agentbridge/controlplane/saasplane` 이다.

이 adapter 는 `github.com/teamswyg/riido-contracts/assignment v0.3.0` 의 shared
DTO/state/event contract 를 사용한다.

It translates SaaS assignment poll/heartbeat/event HTTP API into
TaskSourcePort/TaskReporterPort.

HTTP handler, store actor, SSE, authZ, metrics/health, persistence,
Terraform/AWS/deploy evidence 는 여전히 `riido-control-plane` 또는 `riido-infra` 가
소유한다.
