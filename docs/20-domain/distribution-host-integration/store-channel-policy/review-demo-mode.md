# Review / Demo Mode

[Back to Store Channel Policy](../store-channel-policy.md)

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
