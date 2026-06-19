# Open Issues

[Back to runtime-actor-boundary.md](../runtime-actor-boundary.md)

Open questions roadmap 문서는 [`../50-roadmap/open-questions.md`](../../../50-roadmap/open-questions.md) 가 소유한다.
`Q-RT-001` 과 legacy `Q-MULTICA-005` 는 §7.5 로 닫혔고, `Q-CTX-001` 은 §7.5/§7.7 로 닫혔으며,
`Q-RT-003` 은 §5.6 으로 닫혔고, `Q-RT-005` 는 §8 로 닫혔다.

- `Q-RT-002`: provider process crash 와 lease handoff 사이의 정확한 ordering (`ConnectionLost` draft -> ingest -> handoff orchestration).
- `Q-RT-004`: wrapper 매니페스트의 표준 위치 / 형식(공개 spec vs 사내 전용).
- `Q-RT-006`: Codex app-server `thread/fork` 같은 experimental surface 의 사용 가부. `task.allowExperimentalRuntime` 외에 어떤 추가 게이트가 필요한가.
