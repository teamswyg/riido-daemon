# 8 Workspace Lock Policy

[Back to native config manifest](../native-config-manifest.md)

본 문서는 lock 의 **사용 정책** 을 도메인 표현으로 갖는다. 실제 `flock` /
DB lease primitive 는 [`../locking.md`](../locking.md) (C9).

## 8.1 lock 의 종류

| Lock | scope | 보유 시간 | 사용 시점 |
| --- | --- | --- | --- |
| `repo_cache_update.lock` | `cache/repos/{repo_hash}` 단위 | **짧음** — fetch/prune 동안만 | cache 갱신 시 |
| `task_workdir.lock` | `workspaces/.../runs/{run_id}/` 단위 | run 동안 보유 (로컬 OS 동시성 보호) | adapter / RunController / validation 이 같은 workdir 을 동시 mutate 못하게 |
| `archive_pipeline.lock` | archive 단계 단위 | archive 동안만 | `ArchiveWorkspace` |

## 8.2 금지

- **`repo_lock` 으로 agent run 전체를 감싸지 않는다.** 같은 repo 의 여러 task 가 직렬화되어 처리량이 깨진다.
- **`task_workdir.lock` 으로 다른 task 의 workdir 를 보호하려 하지 않는다.** 격리는 디렉토리 분리 + protected path / sandbox 로 한다.

## 8.3 분산 환경에서

여러 데몬이 같은 호스트 / 같은 cache 를 공유할 때는 `repo_cache_update.lock` 만
의미가 있다. 다른 host 의 데몬과의 cache 공유는 본 SSOT 비범위 — 보통 host 별
독립 cache 를 둔다.
