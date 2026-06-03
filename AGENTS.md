# AGENTS.md

This repository is SSOT-document-first. Before changing behavior or process,
update the owning document under `docs/` in the same PR.

## Work Branch Rule

Every change starts from a Riido task. Use only the `branchName` returned by the
Riido task creation response as the Git branch name.

The SSOT for this rule is
[`docs/30-architecture/riido-work-branch-gate.md`](docs/30-architecture/riido-work-branch-gate.md).
Pull requests that do not follow the Riido work-unit branch shape fail
`.github/workflows/riido-work-branch.yml`.

