# RIID-4570: Store App Repo / Adapter Ownership

[Back to distribution-host-integration](../distribution-host-integration.md)

This slice closes the Store App ownership discussion by moving `Q-CTX-005` out
of open questions and into C11 / architecture SSOT.

Decisions:

- `riido-daemon` owns C11 pure domain facts, helper runtime planning, local IPC
  server contracts, and store distribution gates
- a future desktop/app repository may own concrete Store App GUI, native
  entitlement calls, picker/bookmark adapters, App Store/MSIX project files, and
  submission UI surfaces
- Store App GUI must remain a client of C11/local API contracts and must not
  spawn provider CLIs directly, bundle provider CLIs, or copy C11 domain facts
- signing/provisioning secrets and live store submission evidence remain outside
  public repositories

The slice adds focused public CI that fails if `Q-CTX-005` returns to daemon open
questions or if the Store App ownership SSOT loses repository boundary wording.
