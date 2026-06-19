# Daemon Lifecycle Adapter Scope

[Back to Integration Gates](../integration-gates.md)

RIID-4690 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`cmd/riido daemon ...` lifecycle adapter 다.

이 adapter 는 public provider adapters, `runtimeactor`, `supervisor`,
`taskdbplane`, and `saasplane` 을 하나의 customer-PC process 로 조립한다.

It exposes status/health/ready/metrics JSON on a local-only Unix socket.

Provider CLI binary bundling, server HTTP/SSE implementation,
Terraform/AWS/deploy evidence, and private machine-local state 는 이 context 밖에
남는다.
