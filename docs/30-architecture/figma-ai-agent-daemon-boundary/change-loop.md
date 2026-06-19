# Change Loop

[Back to Figma AI Agent Daemon Boundary](../figma-ai-agent-daemon-boundary.md)

Top-down:

1. Figma or planning changes saved data, generated API, or assignment meaning.
2. `riido-contracts` and `riido-control-plane` SSOT/API DSL change first.
3. daemon updates only when the new meaning reaches assignment snapshot,
   lifecycle command, liveness field, or provider-runtime input.

Bottom-up:

1. daemon runtime/provider/detection harness finds a real constraint.
2. local boundary evidence records the daemon fact first.
3. promote upstream only when client-facing semantics must change.
