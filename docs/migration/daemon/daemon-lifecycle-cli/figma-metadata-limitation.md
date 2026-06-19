# RIID-4843: Figma Metadata Page-List Limitation Guard

[Back to daemon-lifecycle-cli](../daemon-lifecycle-cli.md)

This slice mirrors the upstream `riido-contracts` Figma metadata tooling
limitation into the daemon boundary manifest.

This slice does:

- record `teamswyg/riido-contracts#52` as upstream provenance for
  `figma-metadata-page-list-underreports-pages.v1`
- require the daemon projection to preserve authoritative pages `129:5215`,
  `42:3014`, and `0:1`, even when supporting metadata lists only the UI page
- require non-UI/onboarding daemon evidence nodes to stay in
  `figma-ai-agent-daemon-boundary.riido.json`
- add a focused public Go test for the mirrored limitation and downstream non-UI
  node preservation

This slice does not change daemon runtime behavior, add Figma integration, add
SaaS endpoints, or make daemon the owner of Figma page discovery. The canonical
Figma coverage and inspection method remain in `riido-contracts`; daemon only
guards its downstream projection.
