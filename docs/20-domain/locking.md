# Locking / Lease SSOT

> **이 문서가 C9 Locking / Lease primitive 의 SSOT다.**
>
> - 책임: local file lock primitive, task DB sidecar lease registry, fencing token 증가 / 비교의 infra 규칙.
> - 비책임: 어떤 runtime 이 task 를 잡을 수 있는가의 scheduling 결정은 C5 daemon migration slice 가 소유한다. task 상태 전이는 public [`riido-contracts`](https://github.com/teamswyg/riido-contracts) 의 C1 계약이 소유한다. provider process 실행은 C4 daemon migration slice 가 소유한다.

이 SSOT 는 **C9 Locking / Lease** context 를 채운다.

## 핵심 invariant

1. **C9 는 primitive 만 제공한다.** C9 는 lock 획득 / release, lease sidecar 갱신, fencing token 증가를 보장한다. 어떤 task 가 eligible 한지는 C5 가 결정한다.
2. **local JSON task DB mutation 은 file lock 아래에서만 수행한다.** `riido-task-db.v1`, `riido-runtime-registry.v1`, `riido-runtime-lease-registry.v1` 을 함께 다루는 adapter 는 같은 `.lock` file 을 잡고 읽기-수정-쓰기 순서를 직렬화한다.
3. **fencing token 은 task 별 monotonic counter 다.** active foreign lease 가 있으면 claim 은 실패한다. expired 또는 released lease 를 다시 잡으면 token 은 이전 값보다 1 증가한다.
4. **lease pin 은 C5 값과 일치해야 한다.** lease record 의 `(RuntimeID, CapabilityFingerprint)` 는 C5 `RuntimeLease` 의 pin 이다. fingerprint 가 바뀌면 기존 lease 는 stale 이며 재사용하지 않는다.
5. **file lock 은 local-only primitive 다.** 현재 구현은 같은 host 의 여러 daemon process 를 직렬화한다. 원격 DB / 분산 claim 은 별도 adapter 가 같은 C9 의미를 다른 primitive 로 구현해야 한다.

## Detail Surfaces

- [Local file lock](locking/local-file-lock.md)
- [Local task DB lease registry](locking/lease-registry.md)
- [Acquire / release rules](locking/acquire-release.md)
- [Request metadata](locking/request-metadata.md)
- [Adjacent SSOT contracts](locking/adjacent-contracts.md)
