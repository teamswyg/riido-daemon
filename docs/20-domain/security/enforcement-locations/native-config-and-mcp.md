# 5.2-5.4 Native Config and MCP Lifecycle Enforcement

[Back to enforcement locations](../enforcement-locations.md)

## 5.2 Native config hook materialization 정책

`internal/policy.EvaluateNativeConfigHook` 은 T-CFG provider-native hook surface 의
순수 결정 함수다. 현재 실행 surface 는 Claude Code command hook 설정/스크립트
주입을 audit-only 로 허용하는 `claude:command-hooks:audit` 하나다. C6
`internal/workdir` 는 hook 정책을 결정하지 않고, C4/C6 경계의 supervisor 가 active
policy bundle 을 평가해 `claude-command-hooks` 또는 `instruction-only` hook mode 를
넘긴다.

| Surface | 집행 위치 | 기본 local policy |
| --- | --- | --- |
| Claude audit-only command hooks (`.claude/settings.json`, `.riido/hooks/claude-audit-hook.sh`) | C4 supervisor → `internal/workdir.InjectRuntimeConfig` | Host 에서 허용 |

policy bundle 이 해당 surface 를 허용하지 않으면 Claude provider 여도 `CLAUDE.md`
만 주입하고 `.claude/settings.json` / hook script 는 만들지 않는다. `Unknown`
trust tier 는 항상 `instruction-only` 로 수렴한다.

## 5.3 Native config file materialization 정책

`internal/policy.EvaluateNativeConfigFile` 은 T-CFG provider-native config file/home
surface 의 순수 결정 함수다. 현재 기본 local bundle 은 native config file surface
를 허용하지 않는다. `codex:config-home:task-scoped` 는 과거 compatibility surface 로
남아 있지만 현재 `riido-native-config-plan.v1` 은 Codex `.codex/config.toml` 또는
adapter `CODEX_HOME=<workdir>/.codex` metadata 를 materialize 하지 않는다.

| Surface | 집행 위치 | 기본 local policy |
| --- | --- | --- |
| Codex task-scoped config home (`.codex/config.toml`, adapter `CODEX_HOME`) | compatibility parser surface only | Host 에서 비허용 |

Codex provider 는 항상 `AGENTS.md` 만 native config 로 주입한다. app-server
credential 사용과 full-access sandbox envelope 는 C4 Codex command builder 가
소유한다. `Unknown` trust tier 는 모든 provider config-home materialization 을
차단한다. Codex process 가 실행 중 생성하는 workdir `.codex` state 는 이 표의 허용
대상이 아니며, daemon 은 그 경로를 provider-native config overlay 로 선언하지 않는다.

## 5.4 MCP temp file lifecycle 정책

MCP raw JSON config 는 provider command line 에 직접 inline 하지 않고 adapter 가
소유한 temp file path 로만 전달한다. C4 adapter 는 `StartCommand.TempFiles` 에 이
path 를 보고하고, C4 session actor 는 provider process exit / cancellation / timeout
으로 run 이 종료될 때 해당 temp file 을 삭제한다.

이 규칙은 Claude `--mcp-config` 처럼 provider protocol 이 file path 를 요구하는
경우의 lifecycle gate 다. Temp file 삭제 실패는 run-scope warning 으로만 남기고
terminal result 를 바꾸지 않는다. 같은 path 가 중복 보고되거나 이미 삭제된 경우는
idempotent cleanup 으로 처리한다. Provider adapter 는 temp file path 를 `DroppedArgs`
/ telemetry payload / command warning 에 raw config 내용으로 풀어 쓰면 안 된다.
