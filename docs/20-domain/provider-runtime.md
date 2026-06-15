# Provider Runtime / Adapter SSOT

This SSOT is split into focused parts so handwritten files stay below the repository line-count threshold.
The original entrypoint remains here to preserve existing links.

## Compatibility Markers

- `figma-ai-agent-daemon-boundary`
- `RIID-4901`
- `provider-validation-matrix.riido.json`
- `supports_worktree=false`
- `required_surfaces=[worktree]`
- `MISSING_REQUIRED_SURFACE:worktree`
- Provider full-access/trusted modes are not assumed from provider defaults or
caller arguments
- daemon-owned full-access runtime selection
- Codex adapter 가 danger-full-access envelope 만 생성하고 그 위험을 Riido harness 가
관리한다
- not a provider default, caller-provided default, or
  hidden fallback
- Other providers should follow the same full-access/trusted-runtime
meta model only through provider-specific SSOT

## Parts

- [Part 01: Part 01](provider-runtime/part-01.md)
- [Part 02: 0. Public Migration Status](provider-runtime/part-02.md)
- [Part 03: Part 03](provider-runtime/part-03.md)
- [Part 04: 1. 책임 한 줄](provider-runtime/part-04.md)
- [Part 05: 4.1 허용 필드 (어댑터가 채울 수 있는 것)](provider-runtime/part-05.md)
- [Part 06: 6. raw → draft 변환 규칙 (어댑터 ACL)](provider-runtime/part-06.md)
- [Part 07: 7.7 RuntimeActor boundary](provider-runtime/part-07.md)
- [Part 08: 12. version-affecting changes](provider-runtime/part-08.md)
