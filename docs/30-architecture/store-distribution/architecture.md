# Store Distribution Architecture SSOT: Architecture

[Back to store-distribution.md](../store-distribution.md)

This entrypoint defines the store-distribution architecture surface for Riido
daemon packaging. The executable contract remains
[`packaging/store/riido_daemon_store_distribution.riido.json`](../../../packaging/store/riido_daemon_store_distribution.riido.json)
and is verified by `tools/storecontract`.

Focused sections:

- [Distribution decisions](architecture/decisions.md)
- [Target matrix](architecture/target-matrix.md)
- [MSIX acceptance criteria](architecture/msix-acceptance.md)
- [Mac App Store acceptance criteria](architecture/mac-app-store-acceptance.md)
- [Package boundaries](architecture/package-boundaries.md)
- [macOS helper / login item strategy](architecture/macos-helper-login.md)
- [Windows MSIX runtime / packaging strategy](architecture/windows-msix-runtime.md)
