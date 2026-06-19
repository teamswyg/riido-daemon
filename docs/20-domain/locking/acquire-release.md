# Acquire / Release Rules

[Back to Locking / Lease SSOT](../locking.md)

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
