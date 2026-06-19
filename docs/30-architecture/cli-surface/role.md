# Role

[Back to CLI Surface SSOT](../cli-surface.md)

The CLI may:

- parse args and print usage
- call local daemon packages in this repository
- read/write local JSON state through guarded adapters
- open local IPC transports only
- emit JSON for shell/operator automation

The CLI must not:

- start a public network listener
- bundle or install provider CLIs
- run infrastructure deploy/apply workflows
- own SaaS server behavior
- bypass task mutation guards or policy gates
- redefine contract facts owned by `riido-contracts`
