# Task Source Selection

[Back to Daemon Config Reference](../config-reference.md)

Exactly one production task source may be selected.

| Variable | Consumer | Default | Meaning |
| --- | --- | --- | --- |
| `RIIDO_TASK_QUEUE_DIR` | file queue source | empty | provider-neutral task JSON files |
| `RIIDO_TASK_REPORT_DIR` | file reporter | `RIIDO_TASK_QUEUE_DIR/reports` | JSONL report output |
| `RIIDO_TASK_DB_SOURCE_PATH` | `taskdbplane` | empty | local `riido-task-db.v1` production source |
| `RIIDO_SAAS_URL` | `saasplane` | empty | SaaS assignment polling endpoint |
| `RIIDO_DAEMON_POLL_INTERVAL_SECONDS` | supervisor | `1` | active/fast claim polling interval |
| `RIIDO_DAEMON_IDLE_POLL_INTERVAL_SECONDS` | supervisor | `5` | idle retry interval; must be >= active interval |
| `RIIDO_DAEMON_HEARTBEAT_INTERVAL_SECONDS` | supervisor | `5` | runtime heartbeat interval |

Queue, task DB, and SaaS variables are mutually exclusive where adapters would
otherwise compete for task ownership.
