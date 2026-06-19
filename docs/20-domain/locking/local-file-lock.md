# Local File Lock

[Back to Locking / Lease SSOT](../locking.md)

Go 구현은 `internal/lock` 이 소유한다.

| primitive | 구현 | 의미 |
| --- | --- | --- |
| `AcquireFile(ctx, path)` | `flock(2)` exclusive advisory lock | ctx 가 끝나기 전까지 lock 을 기다린다. |
| `WithFile(ctx, path, fn)` | acquire → `fn` → release | adapter 의 read-modify-write critical section 을 감싼다. |

이 primitive 는 `sync.Mutex` / `sync.RWMutex` 를 쓰지 않는다. 동시성 경계는 OS file lock 이고, actor 내부 상태 보호 수단이 아니다.

`riido-task-db.v1` 에 대한 production adapter 는
`internal/agentbridge/controlplane/taskdbplane` 이며, guarded mutation 자체는
`internal/taskdb` 가 소유한다. `taskdbplane` 만 task DB / runtime registry / lease
sidecar 를 같은 lock 아래에서 함께 read-modify-write 할 수 있다.
