# Windows CLI protection probe

RIID-5971, parent RIID-5900. This is a measurement tool, not a CLI auth helper.
The [custody contract](../../docs/30-architecture/cli-auth-custody.md) remains
authoritative. Windows user authentication is still unavailable.

## What it measures

The probe uses the built-in Microsoft Software Key Storage Provider. It creates
two random, current-user test keys without overwrite, machine scope, existing-key
enumeration or production credentials. All operations that could require UI use
`NCRYPT_SILENT_FLAG`. There is no interactive fallback.

1. The baseline requests nonexportable, signing-only P-256. It checks property
   readback, a real private-export attempt, public-key shape and a real signature.
   Signing is repeated after reopening and in a separate process. Successful
   silent use demonstrates why nonexportability alone is not user presence.
2. The candidate additionally requests strong/high UI protection before
   finalization. A silent finalization refusal is recorded as an interactive
   setup requirement, not proof of protection on later use. If finalization
   succeeds, the probe reads policy back and attempts silent signing through
   the original handle, a reopened handle and a separate process.
3. Only the keys created by this invocation are deleted. Their absence is
   checked before reporting successful measurement. Cleanup failure fails the
   run. No private/public keys, signatures, key names or credentials are printed.

The JSON receipt always has `production_supported: false`. Exit zero means the
measurement completed with its baseline and cleanup intact. Candidate states
such as `silent_use_allowed` are negative product findings, not a failed probe.
Unknown failures remain inconclusive; they do not count as user confirmation.

## Run

Use PowerShell 7 and an isolated disposable Windows test user or the Windows CI job. Do not run
against a user's real CLI installation. From a Visual Studio x64 Native Tools
prompt with the Windows SDK:

```powershell
pwsh -NoProfile -File native/cli-auth-helper-windows/probe/run.ps1
```

The script builds under a fresh temporary directory, runs the silent probe with
a bounded timeout, validates the receipt, writes `out/windows-cli-protection-probe.json`,
and removes its temporary executables. The CI artifact also records runner image
and architecture metadata. This evidence covers the named runner, not all Windows
Desktop versions or NSIS/MSIX packages.

## Remaining acceptance

Interactive setup and the actual kind of prompt, repeated operations before and
after user confirmation, provider cache behavior, credential unwrap, policy
weakening, damaged/replaced keys, publisher/process trust, installer rollback,
and real server authentication are unverified. P-256 is signing-only; no token
storage design is inferred from it. No DPAPI-plus-dialog fallback or new RSA
server wire is introduced. A candidate that permits silent private-key use is
not suitable for the required user-presence boundary.

## Microsoft contracts

- [UI policy](https://learn.microsoft.com/en-us/windows/win32/api/ncrypt/ns-ncrypt-ncrypt_ui_policy)
- [Key properties and export policy](https://learn.microsoft.com/en-us/windows/win32/seccng/key-storage-property-identifiers)
- [Silent finalization](https://learn.microsoft.com/en-us/windows/win32/api/ncrypt/nf-ncrypt-ncryptfinalizekey)
- [Silent signing](https://learn.microsoft.com/en-us/windows/win32/api/ncrypt/nf-ncrypt-ncryptsignhash)
- [Key creation without overwrite](https://learn.microsoft.com/en-us/windows/win32/api/ncrypt/nf-ncrypt-ncryptcreatepersistedkey)
- [Deletion and handle lifetime](https://learn.microsoft.com/en-us/windows/win32/api/ncrypt/nf-ncrypt-ncryptdeletekey)
