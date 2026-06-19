# ExternalToolRegistry

[Back to invariants](../invariants.md)

Provider CLI 등록은 executable path 자체보다 **provenance** 가 중요하다.

현재 순수 도메인 모델은 `internal/hostintegration.ExternalToolRecord` /
`ExternalToolRegistry` 가 실행한다. 이 패키지는 PATH 탐색, provider process
spawn, OS bookmark / named pipe 같은 adapter 일을 하지 않는다. adapter 는 검증된
record 를 이 모델로 넘긴다.

```text
ExternalToolRecord {
    provider              ProviderKind
    executablePath         string
    provenance             "user-selected" | "env-override" | "auto-detected"
    detectedVersion        string
    loginStatus            "unknown" | "logged-in" | "login-required"
    compatibilityStatus    CompatibilityStatus
    lastVerifiedAt         time
}
```

규칙:

1. `user-selected` 가 가장 강한 신호다. Store App 에서 사용자가 file picker 로 지정한 path 다.
2. `env-override` 는 `RIIDO_<PROVIDER>_PATH` 로 들어온 값이다. Store channel 에서는 UI 에 "환경 변수 override" 로 표시해야 한다.
3. `auto-detected` 는 PATH / known install path 탐지 결과다. Store channel 에서는 실행 전 user confirmation 이 필요하다.
4. `login-required` 는 failure 가 아니다. scheduler / UI 가 해당 provider 를 task routing 후보에서 제외할 수 있게 하는 상태다.
