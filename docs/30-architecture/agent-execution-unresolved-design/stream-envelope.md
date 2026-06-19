# Stream Envelope

[Back to Assignment Lifecycle FSM](assignment-lifecycle-fsm.md)

Current partial-body forwarding is a pragmatic daemon-side bridge, but the
server should own user-facing stream semantics.

| Event kind | Store | Client meaning |
| --- | --- | --- |
| `progress_event` | assignment event log | status/progress row |
| `answer_delta` | stream buffer | assistant body live update |
| `final_answer` | terminal assignment result | completed thread body |
| `provider_log` | diagnostics/audit | hidden or expandable diagnostic |

Rule: `final_answer` is the completed body SSOT. `answer_delta` is only for
background correction and live render, and must not replace the terminal result.
