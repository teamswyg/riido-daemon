# Provider Integration Matrix

> Riido task: RIID-4711 `[Daemon] Architecture SSOT docs migration`
>
> This document owns how public `riido-daemon` verifies real provider CLIs.
> Provider CLIs are external attached resources and are never bundled.
>
> Provider-by-provider current evidence is executable in
> [`provider-validation-matrix.riido.json`](provider-validation-matrix.riido.json).

This human-facing matrix is split by verification concern.
The security decision itself is owned by
[`security.md`](../20-domain/security.md) §4.3.

- [Gate policy](integration-matrix/gate-policy.md)
- [Provider matrix](integration-matrix/provider-matrix.md)
- [Integration assertions](integration-matrix/assertions.md)
- [Agent instruction effectiveness probe](integration-matrix/instruction-effectiveness.md)
- [Change procedure](integration-matrix/change-procedure.md)
