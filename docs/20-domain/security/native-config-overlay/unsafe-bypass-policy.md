# Native Config Overlay: Unsafe Permission Bypass Policy

[Back to native-config-overlay](../native-config-overlay.md)

This matrix refines invariant 1 and 2 for `ExposesUnsafePermissionBypass`.

Research history: the matrix is the adopted decision for why Host x bypass is
rejected. Private source research comparing Claude `--permission-mode
bypassPermissions`, Cursor `--yolo`, and Codex unsafe bypass rejection is history;
this matrix is enforcement.

| trust tier x bundle | Behavior |
| --- | --- |
| `Host` x any | always reject with `UNSAFE_BYPASS_ON_HOST` |
| `IsolatedContainer` x bundle allows | allow with single-task isolation and protected path gates |
| `IsolatedContainer` x bundle does not allow | reject |
| `EphemeralVM` x bundle allows | allow |
| `EphemeralVM` x bundle does not allow | reject |
| `CIControlledRunner` x bundle allows | allow after CI isolation guarantee verification |
| `Unknown` x any | always reject |

Current execution meaning of "bundle allows": the provider unsafe bypass surface
is explicitly present in
`bundle.TrustTierPolicies[<tier>].AllowedSurfaces.UnsafeBypass`. There is no
inference and no default allow.
