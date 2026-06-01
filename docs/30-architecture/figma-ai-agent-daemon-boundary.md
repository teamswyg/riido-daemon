# Figma AI Agent Daemon Boundary

> Riido task: RIID-4813 `[Daemon] Figma AI Agent 화면 경계 projection SSOT 게이트`

이 문서는 Figma `v.1.22 AI Agent` 화면을 daemon 관점에서 읽은 projection
입니다. 원본 화면 커버리지의 canonical owner는 `riido-contracts`의
`figma-ai-agent-coverage.riido.json`입니다. 이 문서는 그 결정을 복사하지
않고, daemon이 실제로 실행하거나 소비하는 부분만 좁혀서 고정합니다.

실행 가능한 manifest는
[`figma-ai-agent-daemon-boundary.riido.json`](figma-ai-agent-daemon-boundary.riido.json)
이며 schema는 `riido-figma-ai-agent-daemon-boundary.v1`입니다.

## 판단 기준

- Figma는 제품/디자인 evidence이고, daemon의 durable SSOT가 아닙니다.
- contracts/control-plane이 agent, workspace, thread, generated API 의미를
  먼저 소유합니다.
- daemon은 이미 승인된 assignment, runtime/model/instruction snapshot,
  provider detection/liveness, stop/cancel/lifecycle command만 소비합니다.
- daemon은 client 화면의 copy, sorting, dropdown, modal, scroll, animation,
  timestamp, fixture row, workspace selection, waitlist, marketing consent를
  소유하지 않습니다.

## 주요 화면 경계

| Figma node | 화면 | daemon 판단 |
| --- | --- | --- |
| `153:12742` | 컴포넌트 참여자 드롭다운 | SaaS가 수락한 assignment만 소비합니다. dropdown section, 정렬, row copy는 client/control-plane 경계입니다. |
| `153:15931` | 댓글 소통 | provider progress batch와 stop/cancel 소비가 daemon 경계입니다. 렌더링되는 댓글/thread UI는 client/control-plane 경계입니다. |
| `153:15935` | 추가 기획 내용 | task/subtask assignment만 daemon 실행 입력이 됩니다. project/milestone/intake/mention surface 확장은 daemon이 먼저 만들 수 없습니다. |
| `162:23090` | 런타임 설정페이지 | local current-device lifecycle fact와 수락된 lifecycle command 실행만 daemon 경계입니다. remote read model과 화면 표현은 control-plane/client 경계입니다. |
| `432:37336` | 에이전트 설정페이지 | assigned runtime/model/instruction snapshot만 provider runtime input입니다. agent CRUD, timestamp, editability, list/add affordance는 upstream 경계입니다. |
| `42:3014` / `164:30658` / `435:60050` | 온보딩 | provider detection/liveness evidence만 daemon 경계입니다. workspace 선택, fixture 선택, 직접 설정 form, 설치 CTA, skip modal은 client/control-plane/desktop 경계입니다. |
| `275:22731` | 런타임 설정 empty state | empty state가 device/runtime liveness에서 파생될 수는 있지만 waitlist, install-card hover, marketing consent는 daemon command가 아닙니다. |
| `236:29749` | 웹 온보딩 | download CTA 이후 daemon artifact가 실행될 수는 있어도 sign-up, terms, invite, waitlist는 daemon 경계가 아닙니다. |

## Fixture 용어

Figma와 과거 대화에는 "template"이라는 표현이 섞여 있었지만, 현재 SSOT는
agent template entity를 두지 않습니다. `리도`, `영실`, `홍도`, `지원`은
서버 제공 fixture이며, 선택 결과는 일반 agent 생성으로 이어집니다.

daemon은 fixture catalog, fixture description, fixture instruction copy를
하드코딩하지 않습니다. daemon이 보는 값은 SaaS가 assignment 시점에 확정한
agent instruction/runtime/model snapshot뿐입니다.

## Top-down / Bottom-up Loop

Top-down 변경:

1. Figma 또는 기획이 saved data, generated API, assignment 의미를 바꿉니다.
2. `riido-contracts`와 `riido-control-plane` SSOT/API DSL이 먼저 갱신됩니다.
3. daemon은 새 의미가 assignment snapshot, lifecycle command, liveness field,
   provider-runtime input으로 도착할 때만 실행 경계를 갱신합니다.

Bottom-up 변경:

1. daemon runtime/provider/detection harness가 실제 제약을 발견합니다.
2. 이 문서와 C4/C5/C6/C7 SSOT에 local fact를 먼저 기록합니다.
3. client-facing 의미가 달라지는 경우에만 contracts/control-plane SSOT로
   올려 보냅니다.

## 검증

`go test ./tools/figmaboundary -count=1`은 다음을 확인합니다.

- manifest schema, RIID, Figma file/page identity가 유지되는지
- daemon-relevant Figma node가 모두 entry로 남아 있는지
- 각 entry가 `daemon_scope`, `upstream_owner`, `daemon_consumed_facts`,
  `client_owned_facts`를 구분하는지
- 오래된 agent settings node가 다시 들어오지 않는지
- fixture를 template entity처럼 표현하는 stale 문구가 context/provider-runtime
  문서로 되돌아오지 않는지
- `context-map.md`, `provider-runtime.md`, `daemon.md`, `cli-surface.md`가 이
  manifest를 같은 daemon boundary로 링크하는지
