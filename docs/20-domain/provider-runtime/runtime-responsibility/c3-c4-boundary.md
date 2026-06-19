# C3 To C4 Boundary

[Back to runtime-responsibility.md](../runtime-responsibility.md)

| Question | Owning context |
| --- | --- |
| "What can this provider do?" Surface flags, EventStreamFormat, fingerprint. | **C3 Provider Capability** in public `riido-contracts/docs/20-domain/provider-capability.md` |
| "How does this task run now?" Process start, session resume, stdout parsing, raw to draft. | **C4 Provider Runtime / Adapter** |
| "What domain meaning does this raw event have?" Adapter ACL mapping. | **C4 Provider Runtime / Adapter** |
| "Which runtime owns this task lease?" | **C5 Runtime Scheduling** in `runtime-scheduling.md` |

C4 imports C3 `ProviderCapability` read-only. The reverse direction is forbidden:
public `riido-contracts/provider/capability` must not import daemon runtime
packages.
