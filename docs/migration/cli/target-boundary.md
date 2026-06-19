# Target Boundary

[Back to Riido CLI Migration Plan](../cli.md)

Final CLI boundary:

- `riido mwsd ...`
- `riido task ...`
- `riido serve`
- `riido api ...`
- `riido daemon ...`
- `riido bridge ...`
- local smoke commands that exercise CLI as a black box
- usage/help tests that keep `printUsage()` authoritative

RIID-4685 moved public-safe task/API/bridge wrappers. RIID-4686 restored
`riido mwsd ...` after projection sync was split from private workspace state.
RIID-4690 restored `riido daemon ...` after runtimeactor, supervisor, provider
adapters, taskdbplane, and saasplane became public.

Do not move SaaS server binaries, Terraform/AWS workflows, provider CLI
binaries, private env files, machine-local state, or shared contract facts that
belong in `riido-contracts`.
