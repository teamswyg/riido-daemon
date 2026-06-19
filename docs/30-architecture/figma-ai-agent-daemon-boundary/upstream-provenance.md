# Upstream Provenance

[Back to Figma AI Agent Daemon Boundary](../figma-ai-agent-daemon-boundary.md)

`source_coverage_manifest_provenance.stabilized_by` mirrors the full upstream
coverage history from contracts: `#38`, `#39`, `#45`, `#46`, `#51`, `#52`, `#54`.

`teamswyg/riido-contracts#53` made `stabilized_by` a source field. Daemon records
`mirrors_source_field = "stabilized_by"` and
`source_field_introduced_by = "teamswyg/riido-contracts#53"` so local projection
does not redefine upstream history.

`teamswyg/riido-contracts#54` added node `432:46849`, whose revised onboarding
order is agent draft/configuration -> runtime selection -> workspace selection.
Daemon consumes only the final SaaS assignment after `workspace_id`, `runtime_id`,
instruction, and model are fixed.

The `figma-metadata-page-list-underreports-pages.v1` limitation is local mirror
evidence for RIID-4843. It comes from `teamswyg/riido-contracts#52`.
