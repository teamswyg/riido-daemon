# MWSD And Project Sync Scope

[Back to Integration Gates](../integration-gates.md)

RIID-4686 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/mwsdbridge`, `internal/project`, and `riido mwsd ...` 이다.

`mwsdbridge` 는 macmini-workspace daemon 의 local JSON socket contract 만 읽는
anti-corruption layer 이다.

`project` 는 `riido-workspace-projection.v1` / `riido-project-state.v1` 과
project-to-taskdb projection sync 를 소유한다.

이 sync 는 문서 기반 task source 를 public `internal/taskdb` row 로 투영할 뿐,
provider process execution / runtime session / SaaS transport 를 소유하지 않는다.
