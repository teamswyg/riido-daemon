# Agent Execution Unresolved Design: Overview

[Back to agent-execution-unresolved-design.md](../agent-execution-unresolved-design.md)

Riido task: RIID-4964 `에이전트 프로필을 업로드 하기 위한 endpoint 필요`.

This SSOT captures daemon-side unresolved agent execution risks as executable
architecture inputs. The shared shape is:

1. execution must be keyed by assignment identity, not only logical task id.
2. workspace materialization must be structured data, not prompt text.
3. stop, stream, retry, approval, and resume must converge on one lifecycle FSM.
4. provider launch must freeze executable, PATH, env, cwd, and approval policy.

Focused model files:

- [Problem map](problem-map.md)
- [Current structure evidence](current-structure-evidence.md)
- [Execution identity](execution-identity.md)
- [Workspace plan](workspace-plan.md)
- [Runtime launch envelope](runtime-launch-envelope.md)
- [Assignment lifecycle FSM](assignment-lifecycle-fsm.md)
