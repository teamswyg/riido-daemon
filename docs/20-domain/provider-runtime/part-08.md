# Provider Runtime / Adapter SSOT: Part 08

[Back to provider-runtime.md](../provider-runtime.md)

## 12. version-affecting changes

- `ProviderEventDraft` 의 허용 필드 추가는 `change:additive`.
- 허용 필드 제거 또는 의미 변경은 `change:breaking-protocol` (어댑터 ↔ ingest 와이어 깨짐).
- 금지 필드 표(§4.2)에 새 필드 추가는 `change:breaking-policy` (ingest 계층의 채움 책임을 늘리기 때문).
- `Provider` 포트 시그니처 변경(`StartRun` / `Cancel` / `Drafts()` 등)은 `change:breaking-protocol`.
- 어댑터 4 종 (§10) 중 한 종의 `ProtocolKind` 분류가 바뀌면 public `riido-contracts/docs/20-domain/provider-capability.md` §4 와 동시 갱신.
