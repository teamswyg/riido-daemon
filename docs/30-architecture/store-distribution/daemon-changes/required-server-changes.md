# Required Server Changes

[Back to Daemon Changes](../daemon-changes.md)

| Work | Owner context | Output |
| --- | --- | --- |
| Distribution metadata | C10 | daemon poll/register payload includes channel/app version/status only |
| Provider status sync | C10 | available/login-required/unsupported without path/token |
| Capability routing gate | C10 + C3/C7 | task assignment excludes store-blocked or login-required runtimes |
| Review/demo account | C10 | `review_account_seed.riido.json` store-review-only SaaS seed/provisioning without real provider CLI |
| Privacy policy alignment | C10 | API collection scope matches `privacy_metadata_allowlist.riido.json` public policy metadata allowlist |
