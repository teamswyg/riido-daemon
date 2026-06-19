# Local Task DB Lease Registry

[Back to Locking / Lease SSOT](../locking.md)

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
