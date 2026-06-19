# Migration Order

[Back to Overview](../overview.md)

1. Port SSOT docs first, before moving code that executes decisions.
2. Move provider-neutral primitives: `agentbridge` root types, reducer,
   command/result/event contracts, and concrete-provider-free tests.
3. Move process/workdir/policy/validation support packages behind ports.
4. Move provider adapters one at a time with parser/golden and detect command
   tests. Real CLI integration stays opt-in and skipped unless installed.
5. Move runtime actors and local host integration. Supervisor/runtime/session
   actors remain mailbox-owned, with no shared mutable state shortcut.
6. Rebuild public workflows for unit, domain, generated drift, dependency, and
   black-box daemon checks. Private CI must not duplicate those expensive checks.
