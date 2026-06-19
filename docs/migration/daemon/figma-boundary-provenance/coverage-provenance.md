# RIID-4847 And RIID-4851: Coverage Provenance Mirror

[Back to figma-boundary-provenance.md](../figma-boundary-provenance.md)

These slices tighten the daemon-side Figma boundary provenance guard.
`riido-control-plane` already mirrors the full contracts Figma coverage
stabilization history, but the daemon projection originally recorded only
`teamswyg/riido-contracts#52`.

RIID-4847 expands `source_coverage_manifest_provenance.stabilized_by` to mirror
the full contracts coverage stabilization history:

- `teamswyg/riido-contracts#38`
- `teamswyg/riido-contracts#39`
- `teamswyg/riido-contracts#45`
- `teamswyg/riido-contracts#46`
- `teamswyg/riido-contracts#51`
- `teamswyg/riido-contracts#52`

It keeps `mirrored_supporting_tool_limitations[].source_stabilized_by` narrowed
to `teamswyg/riido-contracts#52`, because that field describes the limitation
slice rather than the whole coverage manifest.

RIID-4851 records that daemon upstream coverage provenance mirrors a
contracts-owned source field:

- `mirrors_source_field = "stabilized_by"`
- `source_field_introduced_by = "teamswyg/riido-contracts#53"`

`tools/figmaboundary` verifies full upstream coverage provenance,
limitation-local provenance, and the source-field marker separately.

These slices do not change daemon runtime behavior, add Figma integration, add
SaaS endpoints, or make daemon the owner of Figma page discovery.
