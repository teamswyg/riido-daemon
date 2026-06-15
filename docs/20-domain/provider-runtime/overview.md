# Provider Runtime / Adapter SSOT: Overview

[Back to provider-runtime.md](../provider-runtime.md)


> **이 문서가 Provider 어댑터 포트 / process · session · run lifecycle / adapter ACL 출력 타입(`ProviderEventDraft`) / cancel · resume · needs-input 처리의 SSOT다.**
>
> - 책임: provider 를 “어떻게 실행하는가” 의 도메인 모델. Provider port interface, process / session / run 의 lifecycle, adapter ACL 변환 규칙, draft 생성 책임.
> - 비책임: provider 가 “무엇을 할 수 있는가” 의 정적 모델 — public `riido-contracts` 의 `docs/20-domain/provider-capability.md` (C3). 어느 task 를 어느 runtime 이 claim 하는가 — [`./runtime-scheduling.md`](./runtime-scheduling.md) (C5). workspace 생성 — [`./workspace.md`](./workspace.md) (C6). 정책 결정 — [`./security.md`](./security.md) (C7). validation 결과 판단 — [`./validation.md`](./validation.md) (C8). **event append authority** — public `riido-contracts` 의 `docs/20-domain/ir-event-log.md` §5.0 와 daemon-side [`internal/ir/ingest`](../../internal/ir/ingest) 구현이 함께 소유한다.

이 SSOT 는 split-repo context map 의 **C4 Provider Runtime / Adapter** context 를 채운다. C1/C2/C3 contract SSOT 는 public `riido-contracts` repository 가 소유하고, 이 repository 는 customer-PC daemon 의 실행 boundary 를 소유한다. C3 ↔ C4 경계는 §2 가 못박는다.
