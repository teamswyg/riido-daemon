# README: Verification

[Back to README](../../README.md)

```bash
go test ./...
go list -m all
go test ./tools/storecontract
go run ./tools/storecontract -contract packaging/store/riido_daemon_store_distribution.riido.json -repo .
go build -o /tmp/riido ./cmd/riido
go run ./cmd/riido --help
go run ./cmd/riido bridge providers
```

`go list -m all` checks the public CI boundary that only Riido-owned modules are
allowed. A new third-party dependency requires a separate decision document and
verification gate first.
