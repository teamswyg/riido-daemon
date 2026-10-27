# CLI protected credential helper

This macOS-only helper is private IPC for the standalone Riido CLI. It does not
require Riido Desktop to be running or signed in. It is not a daemon endpoint and
must never be invoked from an agent run as a silent user-credential fallback.
The owning public contract is `docs/30-architecture/cli-surface.riido.json`.

The helper accepts exactly one bounded JSON document on stdin and returns one
JSON document on stdout. Stdout is a private pipe that may contain credentials;
the parent must never forward it to a terminal, logs, diagnostics, or evidence.
Requests and responses have `version: 1`. Requests contain `operation`
(`status`, `load`, `save`, `delete`, `confirm`), an opaque `key`, and `interactive`. Only
`save` accepts `value`, a base64-encoded opaque blob of at most 32 KiB. The whole
request is at most 64 KiB. Keys are 16–128 URL-safe ASCII characters; the Go
caller owns issuer/account/environment/grant binding and must use a digest.
Storage responses contain `data.present` and, for load only, `data.value`.
`confirm` returns only `data.confirmed: true`; presence metadata, credential
values, false or missing confirmation are never accepted as verification.
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

The `.userPresence` access control protects later reads of stored credentials; first `SecItemAdd` and metadata-only status are not proof of a user-presence check before a fresh OAuth exchange or first API call. Every standalone login now requires `Status -> ConfirmUserPresence -> browser/PKCE -> token exchange -> Whoami -> save`, including account replacement. `confirm` does not access the keychain. It refuses noninteractive input before OS verification and uses a fresh `LAContext` with `.deviceOwnerAuthentication` and no authentication reuse for each request. The bounded wait invalidates the context on completion or timeout. Its fixed `TIMED_OUT` error preserves CLI timeout exit 6; user cancellation remains exit 3 and missing protection exit 8. The parent applies the same two-minute confirmation bound across signing and child execution. Denial, cancellation, timeout, missing support, or invalid response cannot open a browser or replace an existing credential. This explicit OS confirmation source contract still does not prove signed-helper acceptance or server-verifiable installation binding.


### Release bundle metadata

`RIIDO_CLI_RELEASE_VERSION=vX.Y.Z native/cli-auth-helper/build.sh <output-directory> <arm64|x86_64>` embeds the exact release string in `Info.plist` as `RiidoCLIReleaseVersion` **before signing**. Without the variable an unsigned development build uses `development`. Keep the plist in XML format for the portable archive validator. The Apple numeric bundle versions are separate from this exact release binding.

The release packager accepts an already signed bundle through `RIIDO_CLI_AUTH_HELPER_BUNDLE` only on a macOS host targeting Darwin. It copies the complete bundle into a fresh package, checks pinned identity/hardening/entitlements for every Mach-O slice, generates `CLI_AUTH_HELPER.json`, and verifies the unpacked copy again. The current Ubuntu release workflow leaves the input unset and emits legacy archives. No signer, provisioning download, keychain access, notarization, or automatic authentication is added here. Actual signed/profiled acceptance and notarization remain separate held release gates.

### Typed installation key and proof (source implementation)

`installation-key` and `installation-proof` reuse the private version-1 pipe and
dynamic process verification. Both require `interactive: true`. Their `value`
is base64 flat JSON; the outer `key` is the SHA-256 hex of the length-prefixed
domain `riido-cli-installation-key.v1`, canonical API resource and `riido-cli`.
The key request has only `resource` and `client_id`. The proof request additionally
has `installation_intent`, `challenge_hash`, `key_fingerprint`, `code`,
`redirect_uri`, `code_challenge`, `state`, and integral Unix-second `expires_at`.
All fields are mandatory; duplicate, extra, null, noncanonical and expired inputs
are rejected. Constants fix the algorithm, S256 and read scope.

Installation keys use a separate keychain key class/tag from stored OAuth grants.
The helper reuses a permanent P-256 non-extractable key, protected by this-device
user-presence access control. Logout touches only the credential item. The key
operation confirms fresh OS presence before creation/public-key access. Proof
uses a fresh keychain authentication context and verifies its own key fingerprint.
Only public key material and typed proof signatures leave the helper. No generic
signature, private-key export, key deletion or account-triggered rotation exists.
Software keys do not use the Secure-Enclave-only privateKeyUsage flag.

The proof format matches monolith `installation-proof-v1.json`: uint32 big-endian
length prefix per field, domain/authorization_code/resource/client/intent/raw
challenge SHA256/raw key fingerprint/SHA256(code)/redirect/PKCE/S256/SHA256(state)/
read/Unix expiry. Signature is P-256 SHA256 ASN.1 DER, not raw ES256 R||S.
Source fixtures never touch real Keychain or a user dialog. Actual signed helper
acceptance and full CLI-to-server login are separate, unfinished gates.

First-key creation is serialized inside the helper, including direct private-IPC
callers. Keychain ApplicationTag is not a unique primary key: ApplicationLabel
also participates and differs for each random public key. The helper holds an
empty advisory lock under the OS account home (not HOME/TMPDIR), re-reads inside
it, creates at most one key, and verifies the persisted match again. Match-all
queries require exactly one key for reuse/signing; ambiguous state is unavailable
and is never silently repaired or reduced to an arbitrary first match. The lock
checks UID, mode, symlink/hardlink and inode identity, waits at most five seconds,
and releases without deleting the lock file. This coordinates trusted helpers;
same-UID denial of service is outside the user-presence credential boundary.
Official key uniqueness: https://developer.apple.com/documentation/security/ksecclasskey
https://developer.apple.com/documentation/security/ksecattrapplicationlabel
