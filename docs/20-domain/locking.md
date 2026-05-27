# Locking / Lease SSOT

> **이 문서가 C9 Locking / Lease primitive 의 SSOT다.**
>
> - 책임: local file lock primitive, task DB sidecar lease registry, fencing token 증가 / 비교의 infra 규칙.
> - 비책임: 어떤 runtime 이 task 를 잡을 수 있는가의 scheduling 결정은
>   C5 daemon migration slice 가 소유한다. task 상태 전이는 public
>   [`riido-contracts`](https://github.com/teamswyg/riido-contracts) 의 C1
>   계약이 소유한다. provider process 실행은 C4 daemon migration slice 가
>   소유한다.

이 SSOT 는 **C9 Locking / Lease** context 를 채운다.

## 0. 핵심 invariant

1. **C9 는 primitive 만 제공한다.** C9 는 lock 획득 / release, lease sidecar 갱신, fencing token 증가를 보장한다. 어떤 task 가 eligible 한지는 C5 가 결정한다.
2. **local JSON task DB mutation 은 file lock 아래에서만 수행한다.** `riido-task-db.v1`, `riido-runtime-registry.v1`, `riido-runtime-lease-registry.v1` 을 함께 다루는 adapter 는 같은 `.lock` file 을 잡고 읽기-수정-쓰기 순서를 직렬화한다.
3. **fencing token 은 task 별 monotonic counter 다.** active foreign lease 가 있으면 claim 은 실패한다. expired 또는 released lease 를 다시 잡으면 token 은 이전 값보다 1 증가한다.
4. **lease pin 은 C5 값과 일치해야 한다.** lease record 의 `(RuntimeID, CapabilityFingerprint)` 는 C5 `RuntimeLease` 의 pin 이다. fingerprint 가 바뀌면 기존 lease 는 stale 이며 재사용하지 않는다.
5. **file lock 은 local-only primitive 다.** 현재 구현은 같은 host 의 여러 daemon process 를 직렬화한다. 원격 DB / 분산 claim 은 별도 adapter 가 같은 C9 의미를 다른 primitive 로 구현해야 한다.

## 1. Local file lock

Go 구현은 `internal/lock` 이 소유한다.

| primitive | 구현 | 의미 |
| --- | --- | --- |
| `AcquireFile(ctx, path)` | `flock(2)` exclusive advisory lock | ctx 가 끝나기 전까지 lock 을 기다린다. |
| `WithFile(ctx, path, fn)` | acquire → `fn` → release | adapter 의 read-modify-write critical section 을 감싼다. |

이 primitive 는 `sync.Mutex` / `sync.RWMutex` 를 쓰지 않는다. 동시성 경계는 OS file lock 이고, actor 내부 상태 보호 수단이 아니다.

## 2. Local task DB lease registry

`RIIDO_TASK_DB_SOURCE_PATH` 를 쓰는 task DB source 는 task DB 파일 옆에 lease sidecar 를 둔다.

| task DB path | lease registry path | lock path |
| --- | --- | --- |
| `task-db.json` | `task-db.leases.json` | `task-db.json.lock` |

schema version 은 `riido-runtime-lease-registry.v1` 이다.

```json
{
  "schema_version": "riido-runtime-lease-registry.v1",
  "task_db_path": "/path/to/task-db.json",
  "updated_at": "2026-05-25T00:00:00Z",
  "leases": [
    {
      "lease_id": "runtime-lease:task-1:1",
      "task_id": "task-1",
      "runtime_id": "runtime-codex",
      "capability_fingerprint": "sha256...",
      "claimed_at": "2026-05-25T00:00:00Z",
      "lease_until": "2026-05-25T00:00:30Z",
      "fencing_token": 1
    }
  ]
}
```

현재 local JSON lease TTL 은 30초다. TTL 은 provider process 의 최대 실행시간이 아니라 crash window 를 줄이기 위한 local claim fencing window 다. TTL 설정을 env/flag 로 노출할지는 config SSOT 에서 별도 결정한다.

## 3. Acquire / release 규칙

claim adapter 는 같은 file lock 아래에서 다음을 수행한다.

1. runtime registry 를 reload 한다.
2. task DB 를 reload 한다.
3. C5 selector 가 현재 runtime 을 선택했는지 확인한다.
4. task 별 lease record 를 확인한다.
5. active foreign lease 가 있으면 task 를 `Queued` 로 남기고 claim 하지 않는다.
6. lease 가 없거나 expired/released 상태면 새 active lease 를 쓰고 `fencing_token = previous + 1` 로 증가시킨다.
7. same runtime + same capability fingerprint 의 active lease 는 heartbeat 의 `running_task_ids` 에 task id 가 포함될 때 refresh 할 수 있다.
8. lease sidecar 를 저장한 뒤 guarded task transition 을 task DB 에 저장한다.

provider progress / terminal report 가 들어오면 task DB reporter 는 같은 file lock 아래에서 active lease 를 확인한 뒤 C1 transition 을 저장한다. active lease 가 없거나 expired 상태면 progress mutation 은 거절된다. terminal report 는 해당 task lease 의 `released_at` 을 기록한다. provider run 이 끝나면 runtime slot lease 는 반환된 것으로 본다. 이후 validation 은 C8 gate 이며 provider runtime lease 를 계속 잡지 않는다.

claim 전에 task DB source 는 expired lease 를 같은 file lock 아래에서 조정한다. `Preparing` / `Running` task 의 expired lease 는 `BlockerRaised → Blocked` 후 `BlockerResolvedRequeue → Queued` 로 handoff 한다. `Claimed` task 의 expired lease 는 `TaskFailed → Failed` 로 정리한다. `NeedsInput` task 의 expired lease 는 `TaskTimedOut → TimedOut` 으로 정리한다. 이미 waiting / terminal / validation-review 단계로 넘어간 stale lease 는 state 를 바꾸지 않고 release 만 기록한다.

## 4. Request metadata

task DB source 가 claim 에 성공하면 `bridge.TaskRequest.Metadata` 에 다음 값을 넣는다.

| key | 의미 |
| --- | --- |
| `runtime_lease_id` | local lease sidecar 의 `lease_id` |
| `runtime_fencing_token` | task 별 monotonic fencing token |
| `runtime_capability_fingerprint` | C5 selector 가 사용한 capability fingerprint |

supervisor 는 claim metadata 에서 이 값을 typed report context 로 추출해 `StartTask` / `ReportEvent` / `CompleteTask` 호출 context 에 싣는다. task DB reporter 는 active lease 확인과 함께 `runtime_lease_id`, `runtime_fencing_token`, `runtime_capability_fingerprint` 를 sidecar 의 현재 active lease 와 비교한다. 값이 없거나 맞지 않으면 progress mutation 은 거절된다.

## 5. 인접 SSOT 와의 계약

| 인접 context | 본 문서가 받는 / 공급 |
| --- | --- |
| **C5 Runtime Scheduling** | 받는다: selected `(RuntimeID, CapabilityFingerprint)`, lease expiry 의미. 공급: local claim primitive 와 fencing token. |
| **C1 Task Lifecycle** | 받는다: `Queued → Claimed` 등 상태 전이. 공급: 전이를 저장하기 전의 single-writer critical section. |
| **C4 Provider Runtime** | 공급: provider process 가 시작되기 전 claim 이 단일 runtime 으로 fenced 됐다는 사실. |
| **C8 Validation** | provider lease 가 release 된 뒤 검증 command 가 별도 guarded mutation 으로 진행된다는 경계를 공유한다. |
