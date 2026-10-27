# CLI protected credential helper

This macOS-only helper is private IPC for the standalone Riido CLI. It does not
require Riido Desktop to be running or signed in. It is not a daemon endpoint and
must never be invoked from an agent run as a silent user-credential fallback.
The owning public contract is `docs/30-architecture/cli-surface.riido.json`.

The helper accepts exactly one bounded JSON document on stdin and returns one
JSON document on stdout. Stdout is a private pipe that may contain credentials;
the parent must never forward it to a terminal, logs, diagnostics, or evidence.
Requests and responses have `version: 1`. Requests contain `operation`
(`status`, `load`, `save`, `delete`), an opaque `key`, and `interactive`. Only
`save` accepts `value`, a base64-encoded opaque blob of at most 32 KiB. The whole
request is at most 64 KiB. Keys are 16–128 URL-safe ASCII characters; the Go
caller owns issuer/account/environment/grant binding and must use a digest.
Successful responses contain `data.present` and, for load only, `data.value`.
Errors contain only a fixed code and a fixed message; no raw OS error or input.

Every operation performs signature/bundle preflight. The only accepted team is
`3PVV8DQPL6`, the public signing team in riido-desktop `package.json` at main
`5396a5df68ec81685067d06d5ebcf1741e1ead61`. The bundle/signing identifier is
`io.riido.cli-auth-helper`, with dedicated access group
`3PVV8DQPL6.io.riido.cli-auth-helper`. The helper must be packaged in a signed,
hardened-runtime `.app` with an authorized embedded provisioning profile.
Unsigned/ad-hoc/no-profile builds fail closed without contacting the keychain.
A terminal on stdin or stdout is also refused; this is a private-pipe protocol.
The parent must independently verify the helper path, signature, identifier and
team, hardened runtime and absence of unsafe debugging/injection entitlements
*before* sending secrets. A helper cannot authenticate itself to its caller.

Storage uses the data-protection keychain, `.userPresence`, and
`kSecAttrAccessibleWhenUnlockedThisDeviceOnly`. There is no file-based keychain,
plaintext, process-environment or Credential Manager fallback. Noninteractive
requests use an LAContext with interaction forbidden. Status never retrieves
credential bytes. The current implementation has no Windows equivalent; it is
not evidence that Windows or user-presence QA passed.

Apple TN3137 explains why biometric protection needs the data-protection
keychain and app-like packaging for a command-line executable with restricted
entitlements:
https://developer.apple.com/documentation/technotes/tn3137-on-mac-keychains
https://developer.apple.com/documentation/technotes/tn3125-inside-code-signing-provisioning-profiles
https://developer.apple.com/documentation/security/restricting-keychain-item-accessibility

Build: `native/cli-auth-helper/build.sh <output-directory>`. This only compiles an
unsigned development bundle and does not sign, install, or publish it.
Run fixture tests: `native/cli-auth-helper/test.sh`. Tests use fake storage and
never read or write a real keychain. A compiled unsigned helper is also checked
for a sanitized preflight refusal. Signed/profiled synthetic-item QA must still
prove user confirmation, denial, locked state, no-UI refusal, rotation and
removal before the production CLI can claim this storage capability.


### Credential delivery and live process identity

The CLI starts the helper with an empty stdin pipe, keeps the child unreaped, and uses the system `/usr/bin/codesign` to dynamically check that PID with an explicit Apple anchor, Team ID and helper identifier, plus hardened runtime and entitlement policy. Only then does it write any request bytes. Path inspection alone is not a trust boundary in the supported user-writable installation. Failed verification or cancellation closes stdin and terminates/reaps the child. The trusted helper must not exec another program after validation.

A synthetic path-swap regression reproduces the pre-fix credential delivery and now refuses it before any credential bytes are written. This is a negative boundary test, not valid signed-helper acceptance. The current OS's Apple system-process probes did not provide a positive equivalent to a signed Riido helper. Real helper signature/profile/dynamic verification/user-presence QA remains held by the user.

The `.userPresence` access control protects later reads of stored credentials; first `SecItemAdd` and metadata-only status are not proof of a user-presence check before a fresh OAuth exchange or first API call. The browser consent flow, explicit first-login presence and any server-verifiable installation binding must be evaluated separately. This slice does not claim those stronger guarantees.


### Release bundle metadata

`RIIDO_CLI_RELEASE_VERSION=vX.Y.Z native/cli-auth-helper/build.sh <output-directory> <arm64|x86_64>` embeds the exact release string in `Info.plist` as `RiidoCLIReleaseVersion` **before signing**. Without the variable an unsigned development build uses `development`. Keep the plist in XML format for the portable archive validator. The Apple numeric bundle versions are separate from this exact release binding.

The release packager accepts an already signed bundle through `RIIDO_CLI_AUTH_HELPER_BUNDLE` only on a macOS host targeting Darwin. It copies the complete bundle into a fresh package, checks pinned identity/hardening/entitlements for every Mach-O slice, generates `CLI_AUTH_HELPER.json`, and verifies the unpacked copy again. The current Ubuntu release workflow leaves the input unset and emits legacy archives. No signer, provisioning download, keychain access, notarization, or automatic authentication is added here. Actual signed/profiled acceptance and notarization remain separate held release gates.
