# Detect Helper Boundary

[Back to adapter-acl.md](../adapter-acl.md)

`internal/agentbridge/detectutil` 은 concrete provider adapters 가 공유하는 탐지 helper 다. env override 는 hint 가 아니라 pin 이므로 override path 가 없거나 directory 이면 PATH fallback 을 하지 않고 fail-closed 한다.

version probe helper 는 missing binary / timeout / unclassifiable signal 을 unavailable 로 접고, strict probe 는 command completion 여부와 exit code 를 노출해 adapter 가 non-zero output 을 version 으로 오인하지 않게 한다.

`ResolveExecutableCandidates` 는 no-override PATH 후보 목록을 PATH 순서대로 제공하지만, 그 후보를 하나만 쓸지 여러 개 probe 할지는 concrete adapter 의 호환성 정책이다. 현재 OpenClaw 만 calendar-version gate 특성상 지원 버전 후보를 찾을 때까지 later PATH candidate 를 probe 할 수 있다.

`RIIDO_OPENCLAW_PATH` 가 설정된 경우에는 여전히 pin 이며 구버전/오류여도 PATH fallback 을 하지 않는다.

override 가 없을 때 후보 탐색은 process `PATH` 만이 아니라 augmented search path 를 쓴다. Desktop app / launchd / service 로 기동된 daemon 은 최소 `PATH` 만 상속해 Homebrew/per-user 디렉터리에 설치된 `claude`/`codex`/`cursor-agent`/`openclaw` 를 못 찾고 `detection_state=missing` 로 보고할 수 있기 때문이다.

탐색 순서:

1. process `PATH`
2. login-shell `PATH` (프로세스당 1회 `$SHELL -lc` 로 읽어 캐시, Windows / `$SHELL` 미설정 / timeout 시 skip)
3. well-known install directories

이는 unset-override lookup 의 탐색 범위만 넓히며 `RIIDO_<PROVIDER>_PATH` pin 의 fail-closed 의미는 그대로다. 카탈로그는 [`../../30-architecture/config-reference.md`](../../../30-architecture/config-reference.md) 가 소유한다.

Detect 가 선택한 executable path 는 capability snapshot 의 실행 사실이다. `bridge.Run` 과 `runtimeactor.Submit` 은 이 값을 `StartRequest.Executable` 로 `BuildStart` 까지 전달하고, concrete provider adapter 는 이를 다시 `PATH` 에서 재해석하지 않는다.

Adapter specific `StartOptions.Executable` 만 이 값을 override 할 수 있으며, 둘 다 비어 있을 때만 provider default executable name 을 사용한다.
