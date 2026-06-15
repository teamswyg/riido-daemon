# Distribution / Host Integration SSOT: Part 03

[Back to distribution-host-integration.md](../distribution-host-integration.md)

## 6. Store channel policy

| Surface | `developer-id` | `mac-app-store` | `msix-sideload` | `msix-store` |
| --- | --- | --- | --- | --- |
| Provider CLI bundling | 금지 | 금지 | 금지 | 금지 |
| Provider CLI user-selected path | 허용 | 제한 허용: sandbox/security-scoped/user-selected executable grant + Store review approval 필요 | 허용 | 허용 |
| Silent provider auto-install | 금지 | 금지 | 금지 | 금지 |
| Background helper | 허용: 명시 동의 | 제한 허용: login item/helper + sandbox 검토 | 허용: 명시 동의 | 제한 허용: full trust / Store policy 검토 |
| Direct LaunchAgent install | 허용 가능 | 금지 | 해당 없음 | 해당 없음 |
| Windows service install | 해당 없음 | 해당 없음 | 기본 금지 | Store 리스크: 기본 금지 |
| External TCP listener | 금지 | 금지 | 금지 | 금지 |
| Local IPC | 허용 | container/app group 안에서 허용 | 허용 | 허용 |
| Self-updater | 허용 가능 | 금지 | 허용 가능 | 금지: Store 업데이트 우선 |
| Arbitrary home scan | 금지 | 금지 | 금지 | 금지 |

Store policy gate 는 새 runtime security gate 번호를 추가하지 않는다. C11 이 channel 을 판정하고 C7 에 "이 surface 가 이 channel 에서 허용되는가" 를 묻는다. 결과는 C3 compatibility / C4 pre-execute / C6 workspace grant 단계에서 사용된다.

실행 함수는 후속 C7 policy migration slice 의 `internal/policy.EvaluateStoreChannelPolicy` 가 소유한다. 이 함수는 위 표를 그대로 실행하며 provider CLI 실행, helper 설치, IPC listener open, OS entitlement inspection 을 하지 않는다. `dev-local` 은 개발자 배포 실험을 위해 `developer-id` 와 같은 완화 축으로 취급하지만, Provider CLI bundling / silent provider auto-install / external TCP listener / arbitrary home scan 같은 항상 금지 surface 는 동일하게 거절한다.

## 7. Server-facing metadata

Daemon 이 C10 으로 보낼 수 있는 distribution metadata 의 executable model 은
`internal/hostintegration.BuildServerFacingClientMetadata` 다. 이 함수는 local
`ExternalToolRegistry` 를 읽되 C10 에는 routing 에 필요한 최소 status 만 넘긴다.
전송 가능/금지 field 의 실행 가능한 policy artifact 는
`internal/hostintegration/privacy_metadata_allowlist.riido.json`
(`riido-privacy-metadata-allowlist.v1`) 이며, `LoadPrivacyMetadataAllowlist`
와 privacy metadata tests 가 아래 shape 와 artifact 의 일치를 검증한다.

```
ServerFacingClientMetadata {
    distribution_channel
    app_version
    providers[] {
        provider_kind
        provider_available
        provider_login_status
        routing_status
    }
}
```

Field boundary:

| Field | 전송 가능? | 이유 |
| --- | --- | --- |
| `distribution_channel` | 가능 | server routing / store-safe policy |
| `app_version` | 가능 | compatibility / rollout |
| `provider_kind` | 가능 | assignment routing |
| `provider_available` | 가능 | assignment routing |
| `provider_login_status` | 가능 | routing 후보 제외 / UI 안내 |
| `routing_status` | 가능 | C10 sync API / UI / scheduler 공통 vocabulary |
| `provider_executable_path` | 금지 | user filesystem privacy |
| `workspace_root_path` | 금지 | user filesystem privacy |
| provider token / API key | 금지 | secret |

규칙:

1. `distribution_channel` / `app_version` 은 envelope 에 한 번만 담는다.
2. provider list 는 `ProviderKind` 기준 deterministic order 로 만든다.
3. `provider_available=false` 는 실패가 아니라 C10 routing 에서 후보 제외 신호다.
4. `routing_status` vocabulary 는 `available`, `login-required`, `unsupported`, `store-blocked` 네 값만 허용한다. `store-blocked` 는 provider 가 설치되어 있어도 store channel policy 때문에 C10 routing 후보에서 제외되는 상태다.
5. C10 은 이 metadata 를 assignment/capability routing 입력으로만 쓰며, executable path / workspace absolute path / token / API key 를 받거나 저장하지 않는다.
6. full capability fingerprint 나 binary version 은 C3/C4 capability sync 의 별도 계약이 생기기 전까지 이 metadata 에 섞지 않는다.
7. C10 `provider-status` request 에서 받을 수 있는 subset 도 같은 artifact 의 `c10-provider-status-sync-request` surface 가 소유한다. 이 request 는 daemon/runtime identity 와 `distribution_channel`, optional `app_version`, `providers[].provider_kind`, `providers[].routing_status` 만 받으며 `provider_available` / `provider_login_status` 는 C11 projection 내부 field 로만 남는다.

## 8. Review / demo mode

Store 심사용 demo mode 는 다음을 제공한다.

1. provider CLI 미설치 상태 표시.
2. provider 연결 flow 의 read-only preview.
3. workspace grant flow 의 fake/demo root.
4. background helper consent UI.
5. SaaS 없이 local-only status 화면.
6. privacy / telemetry 설정 화면.

Demo mode 는 실제 provider process 를 spawn 하지 않는다. 따라서 C4 Provider Runtime 이 아니라 C11 Store App / Local Helper control surface 의 기능이다.

현재 순수 결정 함수는 `internal/hostintegration.EvaluateReviewDemoMode` 다. 이 함수는 store-managed channel(`mac-app-store`, `msix-store`) 에서 `ConsentLedger` 의 `review-demo-mode` grant 가 있을 때만 review/demo surface 를 활성화한다. 활성화된 demo mode 도 provider execution 과 telemetry sync 를 허용하지 않는다. 즉 reviewer 가 보는 것은 onboarding / provider status preview / workspace grant flow / background consent / privacy setting / local status 화면이며, 실제 C4 provider spawn 과 C10 telemetry sync 는 열리지 않는다.

Store App / helper adapter 가 쓰는 local control surface 는 public `internal/riidoapi` `review-demo` method 다. Request 는 `distribution_channel` 과 `review_demo_consent_granted` 만 받으며, response schema 는 `riido-api-review-demo.v1` 이다. Response 는 `enabled`, `surfaces`, `provider_status_mode`, `provider_execution_allowed=false`, `telemetry_sync_allowed=false`, `local_only=true` 로 reviewer-facing UI 가 어느 화면을 열 수 있는지 알려준다. `riido api review-demo --channel mac-app-store --review-demo-consent-granted true` 는 같은 계약을 CLI 로 검증하는 adapter 이며, provider CLI 를 탐지하거나 실행하지 않는다.

## 9. Open questions

[`../50-roadmap/open-questions.md`](../50-roadmap/open-questions.md) 위임.

- `Q-DIST-001`: RESOLVED by §4.1.1 and §6. `mac-app-store` external provider CLI execution requires a user-selected/sandbox/security-scoped OS grant and Store review approval; otherwise it is store-blocked/demo-only.
- `Q-DIST-002`: RESOLVED by §4.2. `msix-store` 는 packaged full-trust helper/tray process 를 local runtime broker 로 쓰고 Windows service default install 은 금지한다.
- `Q-DIST-003`: ConsentLedger 의 저장 substrate 를 local JSON append log 로 둘지, OS secure storage 를 섞을지.
- `Q-DIST-004`: RESOLVED in the private source open-questions SSOT as `Q-DIST-004` → `Q-CTX-005`. Store App repo ownership 은 C11 context ownership 질문 하나로만 추적한다.

## 10. version-affecting changes

- 새 `DistributionChannel` 추가는 `change:additive` 이지만 StoreChannelPolicy 표 갱신이 필수다.
- provider CLI bundling 금지 invariant 완화는 `change:breaking-policy` 이며 법무/스토어 심사 ADR 없이는 불가하다.
- Local IPC endpoint kind 추가는 C11 additive change 다. C1~C10 이 OS별 IPC 를 import 하게 만들면 context boundary 위반이다.
