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

## Upstream Metadata Limitation Mirror

> Riido task: RIID-4843 `[Daemon] Figma metadata page-list limitation downstream guard`

> Riido task: RIID-4847 `[Daemon] Figma coverage upstream provenance full mirror guard`

> Riido task: RIID-4851 `[Daemon] Figma coverage provenance source-field mirror marker`

daemon manifest의 `source_coverage_manifest_provenance.stabilized_by`는
contracts coverage manifest의 top-level `stabilized_by` source field가 담은
전체 stabilization history
(`teamswyg/riido-contracts#38`, `#39`, `#45`, `#46`, `#51`, `#52`, `#54`)를
미러링합니다. 이것은 daemon이 Figma coverage owner라는 뜻이 아니라,
downstream projection이 어떤 upstream coverage 이력을 기준으로 좁혀졌는지
잃지 않기 위한 provenance guard입니다.

`teamswyg/riido-contracts#54`는 Figma planning node `432:46849`
(`Ex AI - 온보딩 순서 변경 메모`)를 추가했습니다. 이 node의 revised order는
agent draft/configuration -> runtime selection -> workspace selection 이지만,
daemon 관점에서는 client-local draft state가 아닙니다. daemon은 최종
`workspace_id`, `runtime_id`, instruction/model snapshot이 SaaS assignment로
확정된 뒤에만 provider runtime input을 소비합니다. workspace-less create나
persisted draft route는 daemon 실행 경계가 아닙니다.

`teamswyg/riido-contracts#53`에서 contracts manifest가 `stabilized_by`
필드를 직접 소유하게 되었기 때문에, daemon manifest는
`source_coverage_manifest_provenance.mirrors_source_field = "stabilized_by"`와
`source_field_introduced_by = "teamswyg/riido-contracts#53"`도 함께 기록합니다.
즉 daemon은 upstream history를 로컬 기억으로 재정의하지 않고, contracts의
source field를 downstream boundary metadata로 좁혀서 소비합니다.

반대로 `mirrored_supporting_tool_limitations[].source_stabilized_by`는 해당
limitation이 도입된 upstream slice만 기록합니다. 현재
`figma-metadata-page-list-underreports-pages.v1` limitation 자체는
`teamswyg/riido-contracts#52`가 provenance입니다.

`riido-contracts`는 `teamswyg/riido-contracts#52`에서
`figma-metadata-page-list-underreports-pages.v1` limitation을 고정했습니다.
Figma `get_metadata`를 `nodeId` 없이 호출하면 이 파일에서 `129:5215` UI
page만 보일 수 있지만, authoritative page registry는 Figma Plugin API가
로드한 `129:5215`, `42:3014`, `0:1`입니다.

daemon projection은 이 limitation을 downstream에서 미러링합니다. 따라서
metadata page-list 출력은 보조 근거일 뿐이고, 그 결과만 믿어 `42:3014`
온보딩 page나 `0:1` legacy wireframe inventory를 삭제하면 안 됩니다.
특히 `42:3014`, `137:6746`, `138:7389`, `164:26969`, `164:30192`,
`164:30206`, `432:46849`, `435:60050`, `236:29749`, `275:22731` node는 daemon이
직접 실행하지 않는 사실까지 포함해 boundary evidence로 보존되어야 합니다.

## 주요 화면 경계

| Figma node | 화면 | daemon 판단 |
| --- | --- | --- |
| `153:12742` | 컴포넌트 참여자 드롭다운 | SaaS가 수락한 assignment만 소비합니다. dropdown section, 정렬, row copy는 client/control-plane 경계입니다. |
| `153:15931` | 댓글 소통 | provider progress batch와 stop/cancel 소비가 daemon 경계입니다. 렌더링되는 댓글/thread UI는 client/control-plane 경계입니다. |
| `153:15935` | 추가 기획 내용 | task/subtask assignment만 daemon 실행 입력이 됩니다. project/milestone/intake/mention surface 확장은 daemon이 먼저 만들 수 없습니다. |
| `162:23090` | 런타임 설정페이지 | local current-device lifecycle fact와 수락된 lifecycle command 실행만 daemon 경계입니다. remote read model과 화면 표현은 control-plane/client 경계입니다. |
| `432:37336` | 에이전트 설정페이지 | assigned runtime/model/instruction snapshot만 provider runtime input입니다. agent CRUD, timestamp, editability, list/add affordance는 upstream 경계입니다. |
| `42:3014` / `164:30658` / `435:60050` | 온보딩 | provider detection/liveness evidence만 daemon 경계입니다. workspace 선택, fixture 선택, 직접 설정 form, 설치 CTA, skip modal은 client/control-plane/desktop 경계입니다. |
| `432:46849` | Ex AI - 온보딩 순서 변경 메모 | client-local agent draft/configuration 순서 변경은 daemon command가 아닙니다. daemon은 최종 SaaS assignment 이후 runtime/model/instruction snapshot만 소비합니다. |
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
- contracts upstream coverage provenance full mirror와
  `figma-metadata-page-list-underreports-pages.v1` limitation-local provenance
  mirror가 분리되어 유지되는지
- authoritative page `129:5215`, `42:3014`, `0:1`과 non-UI daemon evidence
  node가 metadata 축소 결과로 사라지지 않는지
- daemon-relevant Figma node가 모두 entry로 남아 있는지
- 각 entry가 `daemon_scope`, `upstream_owner`, `daemon_consumed_facts`,
  `client_owned_facts`를 구분하는지
- 오래된 agent settings node가 다시 들어오지 않는지
- fixture를 template entity처럼 표현하는 stale 문구가 context/provider-runtime
  문서로 되돌아오지 않는지
- `context-map.md`, `provider-runtime.md`, `daemon.md`, `cli-surface.md`가 이
  manifest를 같은 daemon boundary로 링크하는지
