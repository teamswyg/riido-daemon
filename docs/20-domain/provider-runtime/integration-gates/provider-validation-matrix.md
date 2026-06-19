# Provider Validation Matrix

[Back to Integration Gates](../integration-gates.md)

RIID-4901 부터 provider별 현재 검증 증거의 executable manifest 는
[`docs/30-architecture/provider-validation-matrix.riido.json`](../../../30-architecture/provider-validation-matrix.riido.json)
다.

이 manifest 는 Claude/Codex/Cursor 의 worktree side-effect PASS 조건과 OpenClaw 의
제한 상태를 분리한다.

OpenClaw 는 text completion, deterministic session id, selected executable evidence 를
가질 수 있지만, C4/C5 runtime capability 는 여전히 `supports_worktree=false` 이다.

따라서 worktree-required task 는 `required_surfaces=[worktree]` 를 통해 C5 scheduling 에서
`MISSING_REQUIRED_SURFACE:worktree` 로 차단되어야 한다.

SaaS completed thread 만으로 filesystem side-effect 를 증명했다고 쓰면 안 된다.
