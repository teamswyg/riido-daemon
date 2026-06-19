# 5.1.2-5.3 Overlay, Deterministic Materialization, and Hook Boundaries

[Back to native config manifest](../native-config-manifest.md)

## 5.1.2 native config overlay policy

Native config overlay 의 표준은 **user-global config 를 읽거나 복사하지 않는
per-task materialization** 이다. Claude command hook 과 future provider-native
config home 은 모두 C7 policy bundle 의 explicit allow surface 로만 활성화된다.
Codex 는 이 C6 overlay 를 쓰지 않고 C4 adapter 의 mandatory full-access sandbox
selection 으로 provider runtime 을 실행한다. Workdir 은 daemon 이 선택한 cwd 와
결과/evidence root 이지만, Codex full-access mode 에서 provider 가 읽고 쓸 수 있는
유일한 filesystem boundary 는 아니다.

OpenClaw / Cursor / unknown provider 는 이 문서 기준에서 instruction-only
overlay 가 default 이며, provider-native config home 을 자동으로 추론하지 않는다.
새 provider-native overlay surface 를 추가하려면 C7 policy bundle surface,
`riido-native-config-plan.v1`, manifest field, NCV 입력을 같은 PR 에서 갱신한다.

## 5.2 deterministic materialization

같은 (`policy bundle version`, `task plan`) 입력은 항상 같은 파일 셋과 같은
내용을 만든다. 임의 timestamp / hostname / random salt 가 파일 내부에 새지 않게
한다. 이게 `NativeConfigVersion` 산출의 기반.

## 5.3 protected path / hooks 의 분리

- “어떤 path 가 protected 인가” = C7 결정 (`T-PATH`).
- “protected 를 어떻게 구현하는가” = C6 (예: chattr `+i`, readonly mount, namespace 격리 등).
- “provider hook 으로 protected path edit 을 차단” = adapter 가 받은 hook script 실행 — C4. hook script 의 내용은 정책 번들에서 옴.
