# 10-11 Resolved Workdir Decisions and Version-Affecting Changes

[Back to native config manifest](../native-config-manifest.md)

아래 항목은 RIID-4573 에서 본 C6 SSOT 로 흡수된 결정이다. 다시 open
question 으로 복제하지 않는다.

| ID | Decision | Follow-up owner |
| --- | --- | --- |
| `Q-WS-001` | Local daemon archive backend default 는 same-host run root `keep-in-place` 다. S3 / 압축 bundle / 외부 storage 는 default 가 아니며, 별도 archive adapter 와 config/env 가 생기기 전에는 자동 선택하지 않는다. | future infra/archive adapter |
| `Q-WS-002` | Default workdir retention 은 disabled 다. `RIIDO_WORKDIR_RETENTION_SECONDS` 가 명시된 경우에만 archived run TTL cleanup 이 켜지고, size / task-count cleanup 은 default 로 존재하지 않는다. | daemon config + workdir cleanup |
| `Q-WS-003` | Shared repo cache prune 은 자동 주기가 없다. 필요 시 operator-triggered maintenance 로만 실행하고 `repo_cache_update.lock` 을 짧게 잡는다. | future cache maintenance adapter |
| `Q-WS-004` | Native config overlay 는 per-task materialization 이 표준이다. User-global config overlay/copy 는 default 로 금지하고, provider-native config home 은 C7 explicit allow surface + manifest/NCV 반영이 있을 때만 쓴다. | C7 policy + C6 workdir |
| `Q-WS-005` | Container/VM workdir 전달 owner 는 C4 runtime launcher / platform adapter 다. C6 는 host-side run root, materialized files, manifest 만 공급한다. | future isolated runtime launcher |
| `Q-WS-006` | Dirty workdir 에 대한 automatic in-place `ReinjectNativeConfig` threshold 는 zero 다. `Preparing`/`Running` 이후 policy/native-config 변경은 runtime-upgrade flow 를 통해 cancel/fail and next-run 재평가로 처리한다. | runtime upgrade flow + supervisor |

## 11. version-affecting changes

- 새 operation 추가는 `change:additive` (단 IR Cat E 이벤트 동시 갱신).
- directory layout 변경은 `change:breaking-policy` + migration tool 필수.
- `NativeConfigVersion` 알고리즘 변경은 `change:breaking-ir` (옛 schemaVersion 산출값을 영원히 보존 — replay 호환).
- lock 정책 변경은 `change:breaking-policy` (분산 환경의 안전성에 영향).
- protected path 구현 방식 변경(예: chattr → namespace) 은 `change:behavioral` (정책 결정은 C7 가 owner, 구현은 본 문서 자유).
