# Local API Adapter Scope

[Back to Integration Gates](../integration-gates.md)

RIID-4684 에서 public `riido-daemon` 으로 이동한 추가 구현 범위는
`internal/riidoapi` local API adapter 다.

이 adapter 는 local IPC envelope 와 Unix-socket / Windows named-pipe transport 를
소유한다.

It calls public `internal/taskdb` guarded mutation and `internal/validation`.

provider runtime 은 이 local API transport 를 소유하지 않는다.
