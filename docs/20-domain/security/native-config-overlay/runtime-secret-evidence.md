# Native Config Overlay: Runtime Secret Evidence

[Back to native-config-overlay](../native-config-overlay.md)

The executable public contract for this boundary is generated from
[`runtime-secret-private-evidence.riido.json`](../../../30-architecture/runtime-secret-private-evidence.riido.json)
and rendered at
[`runtime-secret-private-evidence.md`](../../../30-architecture/runtime-secret-private-evidence.md).

| Evidence kind | Evidence id | Proves | Forbidden |
| --- | --- | --- | --- |
| `runtime-secret-readiness` | `actual-runtime-secret-readiness` | references and payload shape for `RIIDO_AI_SERVER_BEARER_TOKEN`, `RIIDO_AI_SERVER_AUTHZ_TOKENS_JSON`, `RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256` | raw bearer/authZ/review token values |
| `runtime-secret-rotation` | `actual-runtime-secret-rotation` | `rotatable=true`, `last_rotated_at`, `next_rotation_due_at`, and `max_age_seconds` are valid and not due | raw secret values, token values, payload bodies |

Public CI validates that only metadata fields and `ssm:DescribeParameters` are
allowed. Private infra owns any live packet, but that packet must satisfy the
public generated contract and must not contain raw values.
