# Native Config Overlay: Runtime Secret Evidence

[Back to native-config-overlay](../native-config-overlay.md)

`riido_ai_server` production release does not store runtime secret values as
evidence. T-SEC CaaS release evidence is split into two safe evidence kinds.

| Evidence kind | Evidence id | Proves | Forbidden |
| --- | --- | --- | --- |
| `runtime-secret-readiness` | `actual-runtime-secret-readiness` | references and payload shape for `RIIDO_AI_SERVER_BEARER_TOKEN`, `RIIDO_AI_SERVER_AUTHZ_TOKENS_JSON`, `RIIDO_AI_SERVER_REVIEW_ACCOUNT_TOKEN_SHA256` | raw bearer/authZ/review token values |
| `runtime-secret-rotation` | `actual-runtime-secret-rotation` | `rotatable=true`, `last_rotated_at`, `next_rotation_due_at`, and `max_age_seconds` are valid and not due | raw secret values, token values, payload bodies |

`runtime-secret-rotation` is generated from
`riido-runtime-secret-rotation-metadata.v1`.

Rules:

1. Inputs and evidence reject unknown fields fail-closed.
2. Fields such as `value`, `token`, and raw payload bodies fail generation and
   verification.
3. The production SSM Parameter Store slice uses metadata-only
   `aws ssm describe-parameters` JSON and `tools/caasrotationmetadata`.
4. The collector has no `GetParameter`, `GetParameters`, or decrypt path.
5. The collector requires `SecureString`, `Standard` tier, and expected
   parameter names.
6. SSM `LastModifiedDate` is recorded as manual overwrite rotation
   `last_rotated_at`.
7. `next_rotation_due_at` must be after `observed_at`.
8. `last_rotated_at -> next_rotation_due_at` must not exceed each secret's
   `max_age_seconds`.
9. Release packet `apply-ready` mode requires both readiness and rotation
   evidence.
