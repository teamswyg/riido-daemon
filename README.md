# riido-daemon

`riido-daemon` is the public Go module for the Riido local daemon and CLI. It
runs on the user's PC, detects external provider CLIs such as Claude Code,
Codex, OpenClaw, and Cursor Agent, and connects them as executable runtimes
without bundling or installing those CLIs.

This repository is the public/store-reviewable boundary for local helper,
provider connection status, workspace grants, consent, local IPC, and
daemon-side validation.

Read next:

- [Repository boundary](docs/readme/repository-boundary.md)
- [Document map](docs/readme/document-map.md)
- [Module map](docs/readme/module-map.md)
- [Provider CLI principles](docs/readme/provider-cli.md)
- [Run and smoke](docs/readme/run-and-smoke.md)
- [Verification](docs/readme/verification.md)

Module:

```text
github.com/teamswyg/riido-daemon
```

Shared contract dependency:

```text
github.com/teamswyg/riido-contracts v0.3.0
```

License: Apache-2.0.
