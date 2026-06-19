# Validation Gates

[Back to Runtime Snapshot Heartbeat](../runtime-snapshot-heartbeat.md)

Required before a daemon migration PR is mergeable:

```bash
go test ./...
go list -m all
go test ./tools/storecontract
go run ./tools/storecontract -contract packaging/store/riido_daemon_store_distribution.riido.json -repo .
test -f docs/20-domain/context-map.md
test -f docs/30-architecture/config-reference.md
```

When the migrated files include the old audit tooling, restore the stronger
private-repo gates in public CI:

```bash
make check
```

Provider real-CLI integration checks stay environment-gated:

```bash
AGENTBRIDGE_INTEGRATION=1 go test ./internal/provider/... -run TestIntegration -v
```
