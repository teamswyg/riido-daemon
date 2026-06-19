# Figma Task And Runtime Boundaries

[Back to runtime-responsibility.md](../runtime-responsibility.md)

C4 does not interpret task-thread annotations (`node-id=153-15931`) such as
scroll, hover, modal, animation reference, viewer-away state, active stream link
selection, persisted viewer-away visibility, or rendered thread composition.

`riido.aiAgent.events.stream`, `riido.aiAgent.tasks.stop`, and
`riido.aiAgent.tasks.threads` are control-plane/client generated path evidence.
C4 consumes only upstream cancel/interrupt commands and
`<riido_log>{"code":...,"args":{...}}<end>` telemetry markers. Progress code
catalog and append-only policy are upstream `riido-contracts` facts.

C4 does not interpret normal task-thread screen (`node-id=236-21379`) details:
generic comment input, AI Agent reply input, send-button state, right-side task
details panel, or rendered `중지` button. C4 sees no browser click directly;
SaaS polling/assignment response must first deliver cancel/interrupt.

C4 does not interpret participant dropdown annotations (`node-id=153-12742`)
such as member/agent sorting, long-name display, max height, scrollbar width,
or checkbox layout. Assignable-agent response and client composition remain
control-plane/client boundary facts.

C4 does not decide assignment target scope from planning section
`node-id=153-15935`. It does not compute project/milestone/intake/property
filler/mention candidates, AI property filler recommendation exclusion, or
agent mention exclusion. It consumes only SaaS assignments that already passed
target-scope policy.

C4 does not own runtime settings (`node-id=162-23090`) agent hover popover,
daemon stop modal copy, restart animation, remote-device table presentation, or
SaaS device/runtime read model projection. It supplies provider process/run
lifecycle and runtime status only.
