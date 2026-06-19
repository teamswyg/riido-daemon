# Gate Policy

[Back to provider integration matrix](../integration-matrix.md)

Each provider `TestIntegration` is optional until all gates pass:

1. `AGENTBRIDGE_INTEGRATION=1` must be set, otherwise the test skips.
2. The provider executable must be discoverable or explicitly configured with
   `RIIDO_<PROVIDER>_PATH`, otherwise the test skips.
3. The adapter `Detect` result must be available, otherwise the test skips with
   the detect reason.

After all gates pass, a failed prompt roundtrip is a real integration failure.
Provider authentication probes may classify missing login/API-key state as
operator environment skip only when the provider exposes a deterministic probe.

`PASS` in this matrix means the provider produced the evidence named in
`provider-validation-matrix.riido.json`. A skipped integration test, a detected
binary, or a SaaS completed thread alone is not filesystem side-effect evidence.
This is especially important for OpenClaw: its current runtime capability remains
`supports_worktree=false`, so worktree-required tasks must be blocked by the C5
scheduling gate even though OpenClaw can produce text completion and optional
artifact evidence in a locally preconfigured operator environment.

Provider full-access/trusted modes are not assumed from provider defaults or
caller arguments. When Riido chooses such a mode, the provider adapter must make
that launch envelope explicit and the integration evidence must prove the
daemon-selected harness still owns workdir, lifecycle, terminal result, and
filesystem side-effect verification.

The security decision itself is owned by
[`security.md`](../../20-domain/security.md) §4.3; this matrix only records the
provider-specific evidence required to claim that the harness decision is
implemented.
