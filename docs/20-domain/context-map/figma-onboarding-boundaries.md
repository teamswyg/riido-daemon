# Figma Onboarding Boundaries

[Back to context-map.md](../context-map.md)

Figma `node-id=42-3014` onboarding is client/control-plane composition. The
onboarding runtime choice (`node-id=137-6746`), fixture list
(`node-id=138-7389`), direct-setting expansion (`node-id=164-26969`),
workspace scroll affordance (`node-id=164-30192`), selected-workspace and
`새 워크스페이스` rows, two-line ellipsis annotation (`node-id=164-27719`), and
no-installed-AI skip branch (`node-id=164-30206`) are not daemon-owned
decisions.

In `node-id=164-26969`, the `이름`, `설명`, and `지침` input composition maps
upstream to control-plane agent creation; C4 later consumes only the
already-assigned instruction/runtime/model values. Planning node `432:46849`
changes the explanation order to agent draft/configuration, runtime selection,
then workspace selection, but that draft is client-local and does not
authorize a daemon command or workspace-less create path.

`node-id=137-6746` can show Claude Code/Codex as `감지됨` selectable rows and
OpenClaw/Cursor Agent as `감지 안 됨` non-selectable rows. Those labels, radios,
and row states are client presentation over runtime liveness/detection facts.

`node-id=138-7389` can show `리도`, `영실`, `홍도`, and `지원` onboarding fixture
rows, a `직접 설정` row, disabled-next presentation before selection, and a
preview skeleton. These are bootstrap/client composition facts rather than
daemon execution facts.

In the no-installed-AI branch, all-disconnected Claude Code/Codex/OpenClaw/
Cursor Agent rows and the `시작하기` CTA are client presentation over liveness
data, not daemon commands. The daemon supplies runtime liveness/detection facts
and consumes an already-assigned instruction after SaaS authorization. It must
not hard-code onboarding fixture rows, fixture descriptions/instructions,
direct-setting entry points, workspace selection/create-new entry points,
onboarding step skipping, provider install/start CTAs, or client text overflow
behavior.

Figma `node-id=236-29749` web onboarding does not change daemon ownership.
macOS app download is distribution/client routing, not a provider CLI install
command. Google/email sign-up, terms consent, member invite, Windows waitlist,
marketing consent, chat animation, and progress-bar references are
auth/team/product/client facts. The daemon only starts after the desktop/helper
surface launches it and then reports runtime liveness/control-plane assignment
state through the existing SaaS boundary.
