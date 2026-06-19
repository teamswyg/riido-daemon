# Local IPC Rule

[Back to CLI Surface SSOT](../cli-surface.md)

`riido serve`, `riido api`, and `riido daemon` may use:

- Unix socket
- Windows named pipe

They must not add TCP/HTTP listeners to the local CLI binary. SaaS HTTP routes
belong to the public control-plane repository.
