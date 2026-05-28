# riido-daemon

`riido-daemon`은 사용자의 PC에서 실행되는 Riido 로컬 데몬과 CLI를 담는 공개 Go module입니다. Claude Code, Codex, OpenClaw, Cursor Agent 같은 외부 provider CLI를 감지하고 실행 가능한 runtime으로 연결하지만, 그 CLI들을 번들하거나 설치하지는 않습니다.

이 레포는 public/store-reviewable 경계입니다. 즉 App Store, Microsoft Store, Developer ID, MSIX 배포에서 설명 가능해야 하는 로컬 helper, provider 연결 상태, workspace grant, consent, local IPC, daemon-side validation을 공개 문서와 테스트로 검증합니다.

## 이 레포가 하는 일

- `cmd/riido` 단일 CLI/local helper binary를 제공합니다.
- Claude/Codex/OpenClaw/Cursor provider adapter ACL을 구현합니다.
- provider raw stdout/RPC event를 provider-neutral event/result로 정규화합니다.
- runtime actor, supervisor, scheduling, workdir, policy, validation, local task DB adapter를 제공합니다.
- local-only IPC를 제공합니다. macOS는 Unix socket, Windows는 named pipe boundary를 사용합니다.
- Store 심사용 host integration, consent ledger, external tool provenance, review/demo mode 계약을 관리합니다.
- `riido-control-plane` SaaS assignment API를 daemon-side polling/reporting adapter로 소비합니다.

## 이 레포가 하지 않는 일

- provider CLI binary를 포함하거나 자동 설치하지 않습니다.
- SaaS HTTP/SSE server handler, RBAC, control-plane store를 구현하지 않습니다.
- Terraform, AWS, ECR, ECS, release evidence, real deployment config를 소유하지 않습니다.
- App Store/MSIX signing credential, provisioning secret, live submission evidence를 커밋하지 않습니다.
- shared task/IR/provider capability 계약을 다시 정의하지 않습니다. 공통 계약은 `riido-contracts`가 소유합니다.
- public TCP/HTTP listener를 `cmd/riido`에 추가하지 않습니다.

## 왜 이 작업이 daemon 레포에 있는가

AI Agent 기능에서 daemon은 “사용자 기기에서 실제 runtime을 다루는 쪽”입니다. control-plane이 agent/task API를 제공하고, contracts가 DTO와 enum을 고정하고, infra가 배포를 소유한다면, daemon은 다음 결정을 실행합니다.

- runtime(provider CLI)을 외부 도구로 감지하고 연결할 수 있는지
- provider CLI가 없거나 로그인되지 않았을 때 어떤 상태로 보고할지
- 로컬 helper가 어떤 IPC와 app data root를 사용할지
- workspace 접근과 background helper 실행에 어떤 사용자 동의가 필요한지
- Store channel에서 provider 실행이 허용/차단되는 조건이 무엇인지
- task claim, event ingest, validation evidence, local task DB mutation을 어떻게 guarded path로 처리할지

그래서 이 레포의 문서는 단순 개발 메모가 아니라, 스토어 심사와 public CI가 함께 검증해야 하는 daemon-side SSOT입니다.

## 문서 지도

| 알고 싶은 것 | 읽을 문서 |
| --- | --- |
| 전체 문서 입구와 결정 맵 | [`docs/README.md`](docs/README.md) |
| daemon bounded context와 repo 간 책임 분리 | [`docs/20-domain/context-map.md`](docs/20-domain/context-map.md) |
| provider runtime/process/session/adapter ACL | [`docs/20-domain/provider-runtime.md`](docs/20-domain/provider-runtime.md) |
| Store channel, external CLI provenance, consent, local IPC | [`docs/20-domain/distribution-host-integration.md`](docs/20-domain/distribution-host-integration.md) |
| security policy와 store channel execution gate | [`docs/20-domain/security.md`](docs/20-domain/security.md) |
| event redaction과 민감값 처리 | [`docs/20-domain/security-redaction.md`](docs/20-domain/security-redaction.md) |
| runtime scheduling과 eligibility | [`docs/20-domain/runtime-scheduling.md`](docs/20-domain/runtime-scheduling.md) |
| workspace/native config materialization | [`docs/20-domain/workspace.md`](docs/20-domain/workspace.md) |
| CLI command surface | [`docs/30-architecture/cli-surface.md`](docs/30-architecture/cli-surface.md) |
| package/module decomposition과 import rule | [`docs/30-architecture/module-decomposition.md`](docs/30-architecture/module-decomposition.md) |
| App Store/MSIX/Developer ID 배포 기준 | [`docs/30-architecture/store-distribution.md`](docs/30-architecture/store-distribution.md) |
| provider CLI integration 검증 방식 | [`docs/30-architecture/integration-matrix.md`](docs/30-architecture/integration-matrix.md) |
| env var와 flag catalog | [`docs/30-architecture/config-reference.md`](docs/30-architecture/config-reference.md) |
| private `riido_daemon`에서 public daemon으로 옮긴 흐름 | [`docs/migration/daemon.md`](docs/migration/daemon.md) |
| 아직 결정되지 않은 질문 | [`docs/50-roadmap/open-questions.md`](docs/50-roadmap/open-questions.md) |

## 주요 구성

- `cmd/riido`: CLI/local daemon adapter입니다. flag/env를 파싱하고 local-only surface를 조립합니다.
- `internal/agentbridge`: provider-neutral run/request/event/result domain입니다.
- `internal/agentbridge/session`: one-run session actor입니다.
- `internal/agentbridge/runtimeactor`: runtime mailbox, capability reconciliation, slot/heartbeat/stop boundary입니다.
- `internal/agentbridge/supervisor`: task claim, runtime dispatch, workdir preparation, event ingest, result reporting control loop입니다.
- `internal/agentbridge/controlplane/saasplane`: SaaS assignment polling/reporting adapter입니다.
- `internal/agentbridge/controlplane/taskdbplane`: local task DB source/reporter adapter입니다.
- `internal/provider/{claude,codex,openclaw,cursor}`: provider별 external CLI adapter입니다.
- `internal/hostintegration`: Store/host integration pure model입니다.
- `internal/riidoapi`: local IPC API입니다. Unix socket과 Windows named pipe transport를 사용합니다.
- `internal/taskdb`: public daemon copy of `riido-task-db.v1` guarded mutation adapter입니다.
- `internal/mwsdbridge`, `internal/project`: macmini-workspace bridge와 workspace/task projection입니다.
- `packaging/store`, `tools/storecontract`: Store distribution executable contract와 검증 도구입니다.

## Provider CLI 원칙

Claude Code, Codex, OpenClaw, Cursor Agent는 Riido package artifact 안에 들어가지 않습니다. Riido는 사용자가 설치한 외부 CLI를 detect/register/verify할 뿐입니다.

| Provider | 기본 executable | override env |
| --- | --- | --- |
| Claude Code | `claude` | `RIIDO_CLAUDE_PATH` |
| Codex | `codex` | `RIIDO_CODEX_PATH` |
| OpenClaw | `openclaw` | `RIIDO_OPENCLAW_PATH` |
| Cursor Agent | `cursor-agent` | `RIIDO_CURSOR_PATH` |

실제 CLI roundtrip integration test는 opt-in입니다. `AGENTBRIDGE_INTEGRATION=1`이 없거나 provider CLI가 없으면 integration test는 skip해야 합니다. deterministic parser/translator/golden tests는 public CI에서 실행됩니다.

## 실행과 smoke

```bash
go run ./cmd/riido --help
go run ./cmd/riido bridge providers
go run ./cmd/riido daemon start --socket /tmp/riido-agentd.sock --pid-file /tmp/riido-agentd.pid --log-file /tmp/riido-agentd.log
go run ./cmd/riido daemon status --socket /tmp/riido-agentd.sock
go run ./cmd/riido daemon stop --socket /tmp/riido-agentd.sock --pid-file /tmp/riido-agentd.pid
```

`riido daemon ...`은 12-factor env로 task source를 선택합니다.

- `RIIDO_TASK_QUEUE_DIR`
- `RIIDO_TASK_DB_SOURCE_PATH`
- `RIIDO_SAAS_URL` + `RIIDO_SAAS_AGENTS`

이 source들은 서로 경쟁하지 않도록 하나만 production source로 선택해야 합니다.

## 검증

```bash
go test ./...
go list -m all
go test ./tools/storecontract
go run ./tools/storecontract -contract packaging/store/riido_daemon_store_distribution.riido.json -repo .
go build -o /tmp/riido ./cmd/riido
go run ./cmd/riido --help
go run ./cmd/riido bridge providers
```

`go list -m all`은 Riido-owned module만 허용하는 public CI 경계를 확인합니다. 새 third-party dependency가 필요하면 먼저 별도 결정 문서와 검증 gate가 필요합니다.

## Module

```text
github.com/teamswyg/riido-daemon
```

현재 공유 계약 dependency:

```text
github.com/teamswyg/riido-contracts v0.3.0
```

## License

Apache-2.0.
