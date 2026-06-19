# Retired Historical Source Boundary

[Back to Riido CLI Migration Plan](../cli.md)

The initial CLI migration once used the former private source as history. That
source is closed. New work must not read, compare, copy from, cherry-pick from,
push to, open PRs against, merge, or otherwise modify `riido_daemon_private` or
`riido-daemon-private`.

If a CLI fact is missing from public `riido-daemon`, define it in this public
SSOT or promote the shared contract to `riido-contracts`.

Historical CLI slice:

- `cmd/riido`, `printUsage()`, and CLI tests under `cmd/riido`
- scripts that call local CLI commands when not deploy/infra scripts
- README examples for local daemon and local API usage
- `docs/30-architecture/config-reference.md`
- `docs/20-domain/runtime-versioning.md`
- CLI-related roadmap/audit rows
