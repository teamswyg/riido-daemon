# Validation SSOT

> **이 문서가 C8 Validation 의 SSOT다.**
>
> - 책임: `Validating` 상태에서 daemon 이 어떤 검증 사실을 진실로 삼는가, validation command 실행/측정 규칙, `ValidationPassed` / `ValidationFailed` 전이의 의미.
> - 비책임: task FSM 자체와 IR event catalog 는 public
>   [`riido-contracts`](https://github.com/teamswyg/riido-contracts) 의 C1/C2
>   계약이 소유한다. security policy 결정은 C7, provider process 실행은 C4
>   daemon migration slice 가 각각 소유한다.

이 SSOT 는 **C8 Validation** context 를 채운다.

## 0. 핵심 invariant

1. **agent 자기보고는 완료 판정이 아니다.** provider 가 completed result 를 내도 task 는 `Running → Validating` 까지만 간다. `Completed` 는 validation + approval 조건을 통과해야 가능하다.
2. **daemon-measured exit code 가 truth source 다.** validation command 의 성공/실패는 provider/agent 가 보고한 값이 아니라 daemon 이 직접 실행해 관측한 exit code 로 결정한다.
3. **검증 증거는 guarded mutation 으로만 기록한다.** `riido task validate` 와 local API `validate` 는 command 실행 결과를 evidence 로 남긴 뒤, 같은 guarded path 로 `Validating → PatchReady` 또는 `Validating → Failed` 전이를 append 한다.
4. **idempotency 는 command id 가 소유한다.** 같은 command id + 같은 payload 는 기존 receipt 를 반환하고, 같은 id + 다른 payload 는 replay mismatch 로 거절한다.

## 1. Deterministic command gate

현재 구현된 기본 gate 는 `deterministic-command-exit-code.v1` 이다.

| 항목 | 규칙 |
| --- | --- |
| 실행기 | `/bin/sh -lc <command>` |
| 기본 timeout | 5분 |
| workdir | 요청 workdir, 없으면 current working directory |
| exit code `0` | `result=passed` |
| exit code non-zero | `result=failed` |
| timeout | exit code `124`, `result=failed` |
| provider run id | `provider-run:<provider>:<command-id>` |

코드 위치: `internal/validation.RunCommand`.

## 2. Task transition mapping

| validation result | evidence | transition |
| --- | --- | --- |
| `passed` | `ValidationPassed` / deterministic command evidence | `Validating → PatchReady` |
| `failed` | `ValidationFailed` / deterministic command evidence | `Validating → Failed` |

`PatchReady → Completed` 는 validation 의 단독 책임이 아니다. C1 task
lifecycle invariant 에 따라 approval 입력까지 필요하다.

## 3. Security boundary

C7 Security / Policy 는 어떤 validation rule 을 요구할지 결정한다. C8 은 그 rule 을 실행하고 관측 결과를 IR/evidence 로 남긴다. 따라서 C8 은 protected path 목록, secret-scan 정책, diff policy 의 의미를 재정의하지 않는다.

## 4. Version-affecting changes

- gate id 추가는 `change:additive`.
- `deterministic-command-exit-code.v1` 의 exit-code mapping 변경은 `change:breaking-policy`.
- evidence payload schema 변경은 `change:breaking-ir`.
- timeout 기본값 변경은 `change:behavioral`.
