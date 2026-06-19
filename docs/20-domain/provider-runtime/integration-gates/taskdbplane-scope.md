# Task DB Plane Migration Scope

[Back to Integration Gates](../integration-gates.md)

RIID-4683 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/taskdb` 와 `internal/agentbridge/controlplane/taskdbplane` 이다.

`internal/taskdb` 는 `riido-task-db.v1` schema, guarded transition/evidence mutation,
command-id idempotent replay, and deterministic validation evidence receipt 를 소유한다.

`taskdbplane` 은 해당 JSON DB 를 first-class local control-plane source/reporter 로 사용한다.

It performs runtime registry sidecar, lease sidecar, fencing token validation, and expired
lease handoff under the same C9 file lock.

이 slice 는 project/mwsd sync, local API/socket, CLI commands, `saasplane`, server HTTP
transport, infra/secret/state files 를 이동하지 않는다.
