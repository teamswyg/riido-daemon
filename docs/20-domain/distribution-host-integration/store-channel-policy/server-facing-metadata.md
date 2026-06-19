# Server-Facing Metadata

[Back to Store Channel Policy](../store-channel-policy.md)

Daemon 이 C10 으로 보낼 수 있는 distribution metadata 의 executable model 은
`internal/hostintegration.BuildServerFacingClientMetadata` 다. 이 함수는 local
`ExternalToolRegistry` 를 읽되 C10 에는 routing 에 필요한 최소 status 만 넘긴다.
전송 가능/금지 field 의 실행 가능한 policy artifact 는
`internal/hostintegration/privacy_metadata_allowlist.riido.json`
(`riido-privacy-metadata-allowlist.v1`) 이며, `LoadPrivacyMetadataAllowlist`
와 privacy metadata tests 가 아래 shape 와 artifact 의 일치를 검증한다.

```
ServerFacingClientMetadata {
    distribution_channel
    app_version
    providers[] {
        provider_kind
        provider_available
        provider_login_status
        routing_status
    }
}
```

Field boundary:

| Field | 전송 가능? | 이유 |
| --- | --- | --- |
| `distribution_channel` | 가능 | server routing / store-safe policy |
| `app_version` | 가능 | compatibility / rollout |
| `provider_kind` | 가능 | assignment routing |
| `provider_available` | 가능 | assignment routing |
| `provider_login_status` | 가능 | routing 후보 제외 / UI 안내 |
| `routing_status` | 가능 | C10 sync API / UI / scheduler 공통 vocabulary |
| `provider_executable_path` | 금지 | user filesystem privacy |
| `workspace_root_path` | 금지 | user filesystem privacy |
| provider token / API key | 금지 | secret |

규칙:

1. `distribution_channel` / `app_version` 은 envelope 에 한 번만 담는다.
2. provider list 는 `ProviderKind` 기준 deterministic order 로 만든다.
3. `provider_available=false` 는 실패가 아니라 C10 routing 에서 후보 제외 신호다.
4. `routing_status` vocabulary 는 `available`, `login-required`, `unsupported`, `store-blocked` 네 값만 허용한다. `store-blocked` 는 provider 가 설치되어 있어도 store channel policy 때문에 C10 routing 후보에서 제외되는 상태다.
5. C10 은 이 metadata 를 assignment/capability routing 입력으로만 쓰며, executable path / workspace absolute path / token / API key 를 받거나 저장하지 않는다.
6. full capability fingerprint 나 binary version 은 C3/C4 capability sync 의 별도 계약이 생기기 전까지 이 metadata 에 섞지 않는다.
7. C10 `provider-status` request 에서 받을 수 있는 subset 도 같은 artifact 의 `c10-provider-status-sync-request` surface 가 소유한다. 이 request 는 daemon/runtime identity 와 `distribution_channel`, optional `app_version`, `providers[].provider_kind`, `providers[].routing_status` 만 받으며 `provider_available` / `provider_login_status` 는 C11 projection 내부 field 로만 남는다.
