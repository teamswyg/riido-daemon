# ControlPlanePort Boundary

[Back to runtime-actor-boundary.md](../runtime-actor-boundary.md)

`internal/agentbridge/controlplane` 은 daemon supervisor 와 실제 task source/reporter
adapter 사이의 provider-neutral port 계약이다. 이 package 는 "어디에서 task 를
가져오는가" 와 "어디로 결과를 보고하는가" 를 interface 와 local black-box adapter 로
표현하지만, 어떤 runtime 이 선택되는지나 어떤 원격 프로토콜을 쓰는지는 결정하지 않는다.

ControlPlane root package 의 책임:

- `TaskSourcePort`: runtime registration, deregistration, heartbeat, task claim, cancellation watch port.
- `TaskReporterPort`: task start, normalized event, terminal result reporting port.
- claim-time lease metadata 를 reporter 호출 context 로 전달하는 `TaskReportContext` helper.
- `MemorySource` / `MemoryReporter`: tests 와 offline mode 용 RAM-only port implementation.
- `FileQueueSource`: top-level JSON task file 을 atomically claim 하고 claim receipt / runtime registry record 를 남기는 local queue implementation.
- `FileReporter`: task-scoped JSONL report record writer.

ControlPlane root package 의 비책임:

- supervisor polling loop, runtime selection, slot scheduling
- `runtimeactor` session handoff / process execution
- `controlplane/saasplane` HTTP polling / event sync adapter
- task DB source/reporter adapters outside the root package
- `riidoaiserver`, local API, project persistence, packaging, infra, secrets

Public `controlplane/saasplane` owns the SaaS adapter outside the root package.
Public `controlplane/taskdbplane` owns `riido-task-db.v1`, while public
`internal/project` owns project/mwsd projection sync outside this context.
