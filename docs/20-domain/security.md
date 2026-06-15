# Security / Policy SSOT

This SSOT is split into focused parts so handwritten files stay below the repository line-count threshold.
The original entrypoint remains here to preserve existing links.

## Compatibility Markers

### 4.3 Provider full-access runtime harness

Provider full-access runtime harness marker retained for downstream
conformance checks. The policy does not say provider default 가 full-access,
and it does not say caller defaults own the sandbox. Instead, default sandbox 가
danger-full-access 인 provider behavior is rejected as an implicit policy.

Codex adapter 가 danger-full-access launch
envelope 만 생성 when the daemon-owned harness explicitly selects it:

```text
codex --sandbox danger-full-access app-server --listen stdio://
```

daemon 이 Codex 를 전권 host automation surface 로 실행할 때도 harness,
lease, heartbeat, and evidence stay daemon-owned. Claude / Cursor / OpenClaw 도 같은 메타 모델 아래에서 provider-specific trusted-runtime envelope 로만
확장한다.

## Parts

- [Part 01: 0. 핵심 invariant (단단히 박는다)](security/part-01.md)
- [Part 02: 3.1 T-CFG native config overlay decision](security/part-02.md)
- [Part 03: 5.1 코드 집행 위치](security/part-03.md)
