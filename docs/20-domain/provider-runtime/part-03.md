# Provider Runtime / Adapter SSOT: Part 03

[Back to provider-runtime.md](../provider-runtime.md)

A-54 부터 Cursor real CLI integration gate 는 `ResultCompleted` 와 함께 daemon 이
선택한 workdir 안의 expected file artifact 를 확인한다. 이 gate 는
`AGENTBRIDGE_INTEGRATION=1` 과 local Cursor Agent auth/runtime 이 준비된 operator
environment 에서만 실행된다. Cursor adapter 는 native `--workspace <cwd>` 와
`--trust` 를 사용해 daemon-selected workdir 을 전달하며, 이 gate 는 `--yolo` 없이
파일 side-effect 를 확인해야 한다. Gate 가 skip 된 경우에는 filesystem side-effect
가 검증된 것이 아니다.

RIID-4901 부터 provider별 현재 검증 증거의 executable manifest 는
[`docs/30-architecture/provider-validation-matrix.riido.json`](../30-architecture/provider-validation-matrix.riido.json)
다. 이 manifest 는 Claude/Codex/Cursor 의 worktree side-effect PASS 조건과
OpenClaw 의 제한 상태를 분리한다. OpenClaw 는 text completion, deterministic
session id, selected executable evidence 를 가질 수 있지만, C4/C5 runtime capability
는 여전히 `supports_worktree=false` 이다. 따라서 worktree-required task 는
`required_surfaces=[worktree]` 를 통해 C5 scheduling 에서
`MISSING_REQUIRED_SURFACE:worktree` 로 차단되어야 하며, SaaS completed thread 만으로
filesystem side-effect 를 증명했다고 쓰면 안 된다.

RIID-4662 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/agentbridge/supervisor` 다. 이 package 는 Daemon tier control loop 로서
RuntimeActor pool registration / heartbeat, task claim, pre-submit C5 eligibility,
workdir preparation, EventIngestor append delegation, terminal result reporting, and
shutdown cancellation/archive 를 연결한다. RIID-4662 당시에는
`controlplane/saasplane`, `controlplane/taskdbplane`, task DB/project/mwsd/local API,
server HTTP transport, infra/secret/state files 를 후속 migration slice 또는 private
repo 가 맡기로 남겼다.

RIID-4683 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/taskdb` 와 `internal/agentbridge/controlplane/taskdbplane` 이다.
`internal/taskdb` 는 `riido-task-db.v1` schema, guarded transition/evidence
mutation, command-id idempotent replay, and deterministic validation evidence
receipt 를 소유한다. `taskdbplane` 은 해당 JSON DB 를 first-class local
control-plane source/reporter 로 사용하며, runtime registry sidecar, lease sidecar,
fencing token 검증, expired lease handoff 를 같은 C9 file lock 아래에서 수행한다.
이 slice 는 project/mwsd sync, local API/socket, CLI commands, `saasplane`, server
HTTP transport, infra/secret/state files 를 이동하지 않는다.

RIID-4684 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/riidoapi` local API adapter 다. 이 adapter 는 local IPC envelope 와
Unix-socket / Windows named-pipe transport 를 소유하고, public `internal/taskdb`
guarded mutation 과 `internal/validation` 을 호출한다. provider runtime 은 이 local
API transport 를 소유하지 않는다.

RIID-4686 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/mwsdbridge`, `internal/project`, and `riido mwsd ...` 이다.
`mwsdbridge` 는 macmini-workspace daemon 의 local JSON socket contract 만 읽는
anti-corruption layer 이고, `project` 는 `riido-workspace-projection.v1` /
`riido-project-state.v1` 과 project-to-taskdb projection sync 를 소유한다. 이 sync 는
문서 기반 task source 를 public `internal/taskdb` row 로 투영할 뿐, provider process
execution / runtime session / SaaS transport 를 소유하지 않는다.

RIID-4689 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/agentbridge/controlplane/saasplane` 이다. 이 adapter 는
`github.com/teamswyg/riido-contracts/assignment v0.3.0` 의 shared DTO/state/event
contract 를 사용해 SaaS assignment poll/heartbeat/event HTTP API 를
TaskSourcePort/TaskReporterPort 로 번역한다. HTTP handler, store actor, SSE,
authZ, metrics/health, persistence, Terraform/AWS/deploy evidence 는 여전히
`riido-control-plane` 또는 `riido-infra` 가 소유한다.

`saasplane` 은 runtime-snapshot 을 device 단위 full set 으로 보고한다.
`RegisterRuntime` 과 heartbeat refresh 모두 단일 runtime 이 아니라 현재까지 등록된
모든 provider runtime 을 RuntimeID 정렬 순서로 post 한다. 따라서 미탐지 provider 도
`detection_state=missing` 으로 항상 set 안에 남고, snapshot replace 의미를 쓰는 서버
projection 이 device runtime 을 빈 `[]` 로 덮어쓰지 않는다. detected/missing 판정 자체는
runtime capability(`provider.<name>.available`)에서 파생되며, control-plane device
projection(`GET .../ai-agent/devices`)의 최종 표현/필드는 여전히 `riido-control-plane`
이 소유한다.

RIID-4690 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`cmd/riido daemon ...` lifecycle adapter 다. 이 adapter 는 public provider
adapters, `runtimeactor`, `supervisor`, `taskdbplane`, and `saasplane` 을 하나의
customer-PC process 로 조립하고 local-only Unix socket 에 status/health/ready/metrics
JSON 을 노출한다. Provider CLI binary bundling, server HTTP/SSE implementation,
Terraform/AWS/deploy evidence, and private machine-local state 는 이 context 밖에
남는다.

