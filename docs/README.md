# riido-daemon 문서 지도

이 디렉터리는 public `riido-daemon`의 daemon-side SSOT를 담습니다. 코드가 provider 실행과 로컬 helper를 실제로 움직이지만, 어떤 결정이 어떤 맥락에서 내려졌는지는 먼저 이 문서들에서 확인해야 합니다.

## 읽는 순서

1. [`20-domain/context-map.md`](20-domain/context-map.md)에서 daemon이 소유하는 bounded context와 `riido-contracts`, `riido-control-plane`, `riido-infra`와의 책임 경계를 봅니다.
2. [`30-architecture/module-decomposition.md`](30-architecture/module-decomposition.md)에서 package별 역할, import rule, 12-factor daemon boundary를 봅니다.
3. [`20-domain/provider-runtime.md`](20-domain/provider-runtime.md)에서 provider process/session/run lifecycle과 adapter ACL 출력 타입을 봅니다.
4. [`20-domain/distribution-host-integration.md`](20-domain/distribution-host-integration.md), [`30-architecture/store-distribution.md`](30-architecture/store-distribution.md), [`30-architecture/release-artifacts.md`](30-architecture/release-artifacts.md)에서 App Store/MSIX/Developer ID 배포를 위한 role split, provider CLI non-bundling 원칙, GitHub Release binary 경로를 봅니다.
5. [`30-architecture/cli-surface.md`](30-architecture/cli-surface.md)와 [`30-architecture/config-reference.md`](30-architecture/config-reference.md)에서 CLI/env/flag surface를 봅니다.
6. Figma v1.22 AI Agent 화면을 daemon 관점에서 해석해야 할 때는 [`30-architecture/figma-ai-agent-daemon-boundary.md`](30-architecture/figma-ai-agent-daemon-boundary.md)를 봅니다.

## 결정별 문서

| 결정 | SSOT |
| --- | --- |
| public daemon이 어떤 bounded context를 소유하는가 | [`20-domain/context-map.md`](20-domain/context-map.md) |
| `cmd/riido`가 local-only CLI/helper라는 경계 | [`30-architecture/cli-surface.md`](30-architecture/cli-surface.md) |
| provider CLI를 번들하지 않는 이유와 실행 경계 | [`20-domain/distribution-host-integration.md`](20-domain/distribution-host-integration.md), [`30-architecture/store-distribution.md`](30-architecture/store-distribution.md) |
| GitHub Release binary와 curl 설치 경로 | [`30-architecture/release-artifacts.md`](30-architecture/release-artifacts.md) |
| provider process/session/run을 어떻게 domain event draft로 바꾸는가 | [`20-domain/provider-runtime.md`](20-domain/provider-runtime.md) |
| runtime scheduling과 task claim eligibility | [`20-domain/runtime-scheduling.md`](20-domain/runtime-scheduling.md) |
| workspace/native config materialization | [`20-domain/workspace.md`](20-domain/workspace.md) |
| validation evidence와 deterministic command result | [`20-domain/validation.md`](20-domain/validation.md) |
| local lock/lease primitive | [`20-domain/locking.md`](20-domain/locking.md) |
| store channel policy와 tool/security decision | [`20-domain/security.md`](20-domain/security.md) |
| raw provider/tool/event redaction | [`20-domain/security-redaction.md`](20-domain/security-redaction.md) |
| package map과 import rule | [`30-architecture/module-decomposition.md`](30-architecture/module-decomposition.md) |
| Riido 작업 생성 응답의 `branchName`만 쓰는 work-unit branch rule | [`30-architecture/riido-work-branch-gate.md`](30-architecture/riido-work-branch-gate.md) |
| provider CLI real integration test 정책 | [`30-architecture/integration-matrix.md`](30-architecture/integration-matrix.md) |
| runtime upgrade와 compatibility gate | [`30-architecture/runtime-upgrade-flow.md`](30-architecture/runtime-upgrade-flow.md), [`30-architecture/compatibility-gate.md`](30-architecture/compatibility-gate.md) |
| Figma AI Agent 화면에서 daemon이 소비하는 경계 | [`30-architecture/figma-ai-agent-daemon-boundary.md`](30-architecture/figma-ai-agent-daemon-boundary.md) |
| migration history | [`migration/daemon.md`](migration/daemon.md), [`migration/cli.md`](migration/cli.md) |
| 미해결 질문 | [`50-roadmap/open-questions.md`](50-roadmap/open-questions.md) |

## Repo 간 책임 경계

| Repo | 책임 |
| --- | --- |
| `riido-daemon` | customer-PC daemon, CLI/local helper, provider runtime adapters, local IPC, host integration, store distribution executable contract |
| `riido-contracts` | shared task/IR/provider capability/assignment/API contracts |
| `riido-control-plane` | SaaS HTTP/SSE server behavior, RBAC/authZ, assignment store/read model, client-facing API |
| `riido-infra` | Terraform, AWS deployment topology, remote state, private release/deploy evidence |

같은 fact가 두 repo 이상에서 필요하면 daemon 문서에 새로 복사하지 말고 `riido-contracts` 승격을 먼저 검토합니다. 배포-only fact는 `riido-infra`, server runtime fact는 `riido-control-plane`이 소유합니다.

`riido_daemon_private` / `riido-daemon-private`는 retired historical source입니다. 새 작업에서 해당 repository를 참고, 비교, cherry-pick, push, PR, merge 대상으로 사용하지 않습니다. 필요한 결정이나 evidence가 public `riido-daemon`에 없으면 이 문서 디렉터리의 SSOT에 새로 흡수하거나 `riido-contracts`로 승격합니다.

## 작업할 때 지키는 규칙

- 먼저 Riido 작업을 만들고, 반환된 `branchName` 그대로 브랜치를 만듭니다. PR branch가 이 형식을 따르지 않으면 GitHub Actions가 실패합니다.
- `riido_daemon_private` / `riido-daemon-private`를 열람하거나 수정하지 않습니다.
- Provider CLI는 외부 사용자 설치 도구입니다. package artifact, store artifact root, 테스트 fixture에 CLI binary를 넣지 않습니다.
- `cmd/riido`는 local-only입니다. Unix socket 또는 Windows named pipe만 열고 public TCP/HTTP listener를 만들지 않습니다.
- 새 env var나 flag는 [`30-architecture/config-reference.md`](30-architecture/config-reference.md)와 parser test를 같은 PR에서 갱신합니다.
- provider adapter behavior가 바뀌면 deterministic test와 [`30-architecture/integration-matrix.md`](30-architecture/integration-matrix.md)를 함께 갱신합니다.
- store distribution surface가 바뀌면 [`packaging/store/riido_daemon_store_distribution.riido.json`](../packaging/store/riido_daemon_store_distribution.riido.json), [`30-architecture/store-distribution.md`](30-architecture/store-distribution.md), [`30-architecture/release-artifacts.md`](30-architecture/release-artifacts.md), `tools/storecontract` 검증을 함께 확인합니다.
