# RIID-5827 Device assignment notifications — delivery contract

## Acceptance and boundaries
- One authenticated SSE request per daemon/device/environment, regardless of agent count. Idle notification-capable daemons do not call per-agent poll or agent-bindings periodically.
- Notification is an advisory wakeup only. The existing authorized /poll endpoint owns claim/lease, duplicate suppression, resume and cancellation; heartbeat and event/terminal retry remain independent.
- Initial connection and every reconnect reconcile all current device bindings. Changes wake only affected agents. Newly bound agents become visible; removed bindings cannot authorize a claim.
- Store notifications provide immediate local wakeups; a bounded server-side reconciliation interval (5 seconds) detects changes committed on another instance, missed/coalesced signals, lease expiry and binding changes. Only semantic changes cross the stream; heartbeat timestamps do not create notifications.
- Device credentials are checked before subscribing and at each reconciliation; no bearer-only cross-device stream, prompts or assignment content on the stream.
- Legacy daemon /poll remains unchanged. New daemon falls back only when the server explicitly does not implement the endpoint (404/405/501). Network/auth failures use bounded reconnect backoff and cannot cancel a running provider.
- No running daemon restart, task cancellation or active assignment mutation during rollout. Server rolling deployment precedes daemon publication. Existing processes adopt the new daemon only at a safe idle/restart boundary; verify updater behavior before publication.
- Preserve currently deployed production baselines and existing staging commits; integrate only this delta. Do not reset staging to main. Keep the temporary WAF exception until measured adoption makes rollback safe.

## Wire contract
GET /v1/daemon/assignment-events uses existing X-Riido-Device-Id and X-Riido-Device-Secret headers. Content-Type: text/event-stream. Event `assignments` has schema_version `riido-daemon-assignment-events.v1`, bindings (existing agent runtime binding shape), connection_revision, connected_principal_count, and agent_ids to reconcile. Initial/reconnect events include every binding. Subsequent events include affected agents only. Empty bindings are a valid snapshot. SSE comments every 15 seconds maintain liveness. Streams rotate after five minutes; reconnect always re-reads current state, so replay IDs are not required. A lost hint cannot change assignment ownership.

## Safe update adoption
Deployed desktops stop the old daemon before activating its replacement, so daemon-side changes alone cannot protect these clients. Freeze legacy CDN paths `releases/latest/ai-agent` and `releases/staging/ai-agent`. Publish this release only to `releases/safe/ai-agent` (production) and `releases/staging-safe/ai-agent` (staging); the desktop safety PR selects these paths.

The new desktop asks local IPC `shutdown-if-idle` before replacing either a standalone archive or helper bundle and confirms process exit. The supervisor atomically checks preparation, active execution, verification, pending terminal/event delivery and detached reports, then fences new claims before accepting shutdown. Busy, unreachable and legacy daemons (unknown method) defer installation with the old binary/version/ETag intact. There is no automatic PID/signal fallback. Legacy processes adopt only after a natural exit. Explicit operator force-stop retains termination behavior. Publication is not evidence of fleet adoption; desktop review/release and device adoption remain separate gates.
