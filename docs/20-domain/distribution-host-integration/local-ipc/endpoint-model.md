# Local IPC Endpoint Model

[Back to local-ipc.md](../local-ipc.md)

도메인 포트는 다음 semantic 만 노출한다.

```text
LocalIPCEndpoint {
    channel      DistributionChannel
    hostOS       "darwin" | "windows"
    endpointKind "unix-socket" | "named-pipe"
    path         string
    owner        "store-app" | "helper"
}
```

현재 순수 모델은 `internal/hostintegration.LocalIPCEndpoint` /
`DefaultLocalIPCEndpoint` 가 실행한다. 이 모델은 endpoint descriptor 만 계산한다.
실제 listener adapter 는 public `internal/riidoapi` 의 local API transport 가 소유한다.

Channel 별 adapter:

| Channel | IPC adapter | 경로 owner |
| --- | --- | --- |
| `developer-id` | Unix domain socket | `~/Library/Application Support/riido` 또는 user config dir |
| `mac-app-store` | Unix domain socket inside app group/container | App Sandbox container / app group |
| `msix-sideload` | Windows named pipe | package local data / app identity |
| `msix-store` | Windows named pipe | package local data / app identity |
| `dev-local` | Unix domain socket | 현재 launchd/dev path |

`cmd/riido` 의 Unix socket API 는 `dev-local` / `developer-id` adapter 로 본다.
Windows named pipe API 는 `--transport windows-named-pipe` 로 선택하는 C11 local
transport adapter 이며, 같은 request envelope 와 `riidoapi` handler 를 재사용한다.

C1~C10 은 OS별 listener 를 import 하지 않는다. 현재 `cmd/riido daemon` 의 dev-local
기본 socket path 는 C11 `AppDataRoot` + `LocalIPCEndpoint` 를 통해 기존
`$HOME/Library/Application Support/riido/agentd.sock` 로 계산한다. HTTP listener 추가는
여전히 금지다.

RIID-4684 에서 이 adapter 는 private `project/mwsd` 의존성 없이 public
`internal/taskdb` guarded mutation 을 사용하도록 이동됐다.
