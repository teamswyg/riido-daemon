# 5.1 Native Config Manifest and Materialization

[Back to native config manifest](../native-config-manifest.md)

Provider 별 native config file plan 의 실행 가능한 SSOT 는
`internal/workdir/native_config_plan.riido.json`
(`riido-native-config-plan.v1`) 이다. 이 IR 은
`tools/riidogen/templates/native_config_plan.go.gotmpl` 로
`internal/workdir/native_config_plan_gen.go` 를 생성하며,
`go generate ./internal/workdir` 로 갱신한다. 코드가 직접
provider→filename/hook/config-home mapping 을 재정의하면 안 된다.

`riido-native-config-manifest.v1` 은 provider-native config 작성의 현재
증적이다. 이 manifest 는 `workdir/.riido/native-config-manifest.json` 과
`native-config/.riido/native-config-manifest.json` 에 같은 내용으로 쓰인다.

필드:

| Field | 의미 |
| --- | --- |
| `schema_version` | 항상 `riido-native-config-manifest.v1` |
| `provider_kind` | provider family (`claude`, `codex`, `openclaw`, `cursor`, unknown fallback 등) |
| `protocol_kind` | C3 가 선택한 protocol kind. 비어 있으면 field 자체를 생략한다 |
| `primary_instruction_file` | provider 가 자동 로드하는 1차 instruction 파일 (`CLAUDE.md`, `AGENTS.md`, `GEMINI.md`) |
| `manifest_file` | manifest 자체의 상대 경로. 현재 `.riido/native-config-manifest.json` |
| `hook_mode` | 현재 hook materialization 방식. `instruction-only` 는 provider-native hook script/settings 를 아직 쓰지 않고 1차 instruction file hard rule 로만 집행한다는 뜻이고, `claude-command-hooks` 는 Claude `.claude/settings.json` 과 command hook script 를 주입했다는 뜻이다 |
| `config_home_dir` | provider 전용 config home 이 task-scoped 로 주입될 때의 상대 경로. 현재 기본 plan 에서는 비어 있다 |
| `provider_settings_files` | provider 가 직접 읽는 settings/config 파일 목록. 현재 Claude `.claude/settings.json` 을 쓸 수 있다 |
| `hook_files` | provider-native hook settings 가 참조하는 script 파일 목록. 현재 Claude command hook script 는 `.riido/hooks/claude-audit-hook.sh` |
| `telemetry_contract_placement` | SaaS source 가 prompt/system prompt 에 telemetry contract 를 둔 위치. 비어 있으면 field 자체를 생략한다 |
| `workflow` | runtime config workflow branch. 비어 있으면 `default` |
| `generated_files` | C6 가 이 run 에 deterministic 하게 쓴 상대 파일 경로 목록 |

`generated_files` 에는 manifest 자신도 포함된다. 따라서 manifest schema,
hook mode, telemetry placement, provider filename catalog 변경은 모두
`NativeConfigVersion` 에 반영된다.

현재 `riido-native-config-plan.v1` materialization:

| Provider | 생성 파일 | 의미 |
| --- | --- | --- |
| Claude | `CLAUDE.md`, `.claude/settings.json`, `.riido/hooks/claude-audit-hook.sh`, `.riido/native-config-manifest.json` | Claude Code 의 project settings hook surface 를 task workdir 안에 고정한다. 기본 hook 은 `PreToolUse` / `PostToolUse` 입력 JSON 을 `.riido/hooks/claude-hook-events.jsonl` 로 append 하는 audit-only command hook 이며, exit 0 으로 provider 행동을 차단하지 않는다. 단 `.claude/settings.json` 과 hook script 는 C7 policy bundle 이 `claude:command-hooks:audit` surface 를 허용한 경우에만 materialize 된다. 거절되면 manifest 의 `hook_mode` 은 `instruction-only` 로 기록되고 `CLAUDE.md` 만 남는다. |
| Codex | `AGENTS.md`, `.riido/native-config-manifest.json` | Codex 는 task-scoped `.codex/config.toml` 또는 `CODEX_HOME` overlay 를 materialize 하지 않는다. app-server credential 사용과 full-access runtime envelope 는 C4 Codex adapter 가 `codex --sandbox danger-full-access app-server --listen stdio://` 로 고정한다. Codex process 가 실행 중 workdir `.codex` state 를 만들 수 있지만, C6 manifest/provider settings output 으로 선언하지 않는다. Workdir 은 기본 cwd/evidence root 이며 filesystem sandbox boundary 가 아니다. |
| OpenClaw / Cursor / unknown | `AGENTS.md`, `.riido/native-config-manifest.json` | 현재는 provider-neutral instruction file 주입만 한다. |
