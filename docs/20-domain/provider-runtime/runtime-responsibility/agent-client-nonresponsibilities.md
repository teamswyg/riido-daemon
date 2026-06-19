# Agent And Client Non-Responsibilities

[Back to runtime-responsibility.md](../runtime-responsibility.md)

C4 Provider Runtime does not own agent record management or client presentation.

It does not:

- create, save, or update agent settings
- define agent profile / description / instruction meaning or API shape
- stamp agent list `created_at` / `updated_at`
- decide add-screen save enablement
- own row/meatball edit entry
- own no-description row layout, status-label copy/color, long-description presentation, or absolute-time tooltips
- interpret Figma menu placement (`node-id=156-19307`) or route selected state

Agent profile / description / instruction meaning and API shape are upstream
contracts/control-plane facts. C4 consumes only the already-assigned run input
and converts runtime binding, instruction, and model values into provider
process arguments.

Figma menu placement is not a runtime execution input. C4 consumes only a run
after route/client/control-plane authorization has produced an assignment.
