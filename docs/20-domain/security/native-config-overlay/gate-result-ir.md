# Native Config Overlay: Gate Result IR

[Back to native-config-overlay](../native-config-overlay.md)

| Result | EventType | Producer |
| --- | --- | --- |
| gate pass | no `PolicyViolationDetected(...)`; next step proceeds | none; progress itself is evidence |
| gate fail | `PolicyViolationDetected(category, subject, severity)`, then `BlockerRaised(category=POLICY_*)` | caller context |
| policy bundle switch | `PolicyBundleSwitched(from, to)` | C7 |
| scoped token issue | `SecretsScopeIssued(scopeID, ttl, purpose)` | C7 |
| scoped token revoke | `SecretsScopeRevoked(scopeID, reason)` | C7 |

Formal value and payload catalogs are owned by public `riido-contracts/ir` C2
event catalog.
