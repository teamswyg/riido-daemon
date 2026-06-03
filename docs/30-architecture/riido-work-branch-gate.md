# Riido Work Branch Gate

> This document owns the public `riido-daemon` work-unit branch rule.

## Rule

Every code or documentation change must start by creating a Riido task. The
`branchName` returned by that Riido task creation response is the only branch
name that may be used for the work.

Do not invent helper branch names such as `codex/...`, `feature/...`, or
hand-written English slugs. Do not rename a Riido branch locally. If the work
changes enough that the branch name is no longer meaningful, create or update
the Riido task first and use the Riido-provided branch name.

## Why

The Riido task is the work-unit SSOT. A branch without the Riido task key loses
the link between purpose, branch, PR, comments, verification evidence, and
follow-up work.

## Enforced Shape

GitHub Actions enforces the branch name shape on every pull request:

```plain text
<PROJECT_KEY>-<NUMBER>-<SLUG>
```

Examples:

```plain text
A-40-AI-Agent-SSOT-Riido-작업-branchName-사용-강제
RIID-4892-public-provider-smoke-harness
```

The check intentionally rejects namespaced local helper branches such as
`codex/foo`. The format gate proves that a PR branch has Riido work-unit shape.
The exact source of truth for the branch name remains the Riido task creation
response.

## CI Gate

The executable gate is:

```bash
scripts/verify-riido-work-branch.sh "$GITHUB_HEAD_REF"
```

The workflow is `.github/workflows/riido-work-branch.yml`. If the source branch
does not match the Riido work-unit shape, the PR is not mergeable.

