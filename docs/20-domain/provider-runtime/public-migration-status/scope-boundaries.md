# Scope Boundaries

[Back to Public Migration Status](../public-migration-status.md)

RIID-4651 moved `internal/agentbridge` into public `riido-daemon`. This
stdlib-only provider-neutral root includes `Adapter`, `RawEvent` / `Parser`,
`RunState`, reducer, telemetry parser, and tool start gate.

Not included in that slice: task DB, project, MWSD, local API, server,
control-plane, infra, secrets, or state files.

Figma onboarding planning screens are outside C4 ownership. C4 may report
runtime detection/liveness used by onboarding runtime choice and may later
execute SaaS-assigned instruction. It does not own labels, radio state, row
dimming, fixture catalog, direct-setting form, workspace selector, skip branch,
scroll affordance, ellipsis behavior, fixture copy, or preview popover.

Figma node `432:46849` may ask clients to collect agent draft/configuration
before runtime and workspace selection. C4 still starts no provider from that
draft; it consumes only final SaaS-authorized runtime/model/instruction snapshot.

Web onboarding node `236:29749` is also outside C4 ownership. Download CTA may
lead to daemon artifact execution, but sign-up, terms, invite, waitlist,
marketing consent, and progress references stay client/auth/team/product facts.

The daemon projection is
[`../../../30-architecture/figma-ai-agent-daemon-boundary.md`](../../../30-architecture/figma-ai-agent-daemon-boundary.md).
