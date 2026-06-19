# Figma Onboarding And Settings Boundaries

[Back to runtime-responsibility.md](../runtime-responsibility.md)

C4 does not interpret runtime settings empty state (`node-id=275-22731`) provider
install-card hover, Windows app waitlist copy, or marketing-consent state. It
does not bundle, download, or install provider CLIs.

C4 does not interpret web onboarding (`node-id=236-29749`) macOS app download
CTA, sign-up/terms/member-invite flows, Windows waitlist/marketing consent,
chat animation, or progress-bar reference. Auth, team, distribution, and
presentation facts remain upstream.

C4 does not interpret agent settings (`node-id=432-37336`), agent add
(`node-id=134-6542`), agent list/add affordance (`node-id=337-24001` /
`node-id=337-24013`), or agent list (`node-id=432-35713`) create/update form,
save/add-button enablement, "모든 멤버가 런타임이 없으면" presentation, row
edit/delete entry, created/update date stamping, tooltip/layout/copy/color, or
model dropdown catalog.

Runtime-scoped `model_id` may be consumed as an upstream assignment/configuration
input. Provider-specific model catalog and label, and omitted `model_id` default
rules, are owned by public contracts `runtime_model_catalog.v1` and
control-plane read models.

C4 does not interpret onboarding direct-setting expansion (`node-id=164-26969`)
`이름`, `설명`, `지침` form composition, placeholder copy, dimmed fixture rows, or
scroll behavior. It consumes only the instruction/runtime/model values from a
created and assigned agent configuration.
