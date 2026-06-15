# Security / Policy SSOT: Part 01

[Back to security.md](../security.md)


> **이 문서가 trust tier / policy bundle / unsafe bypass 정책 / protected path / network·secrets·destructive action 정책 / 6 security gate / store channel policy decision 의 hub SSOT다.**
>
> - 책임: 무엇이 “허용” 인가, 어느 trust tier 에서 어떤 surface 가 활성 가능한가, 정책 번들의 구조와 진화, 6 security gate 의 위치·검사·실패 동작.
> - 비책임: 정책 **실행** 은 각 인접 context 가 한다. 본 문서는 **결정** 만 한다. workdir 파일 주입 — [`./workspace.md`](./workspace.md) (C6). validation rule 결과 해석 — [`./validation.md`](./validation.md) (C8). security-compatible runtime 배정 — [`./runtime-scheduling.md`](./runtime-scheduling.md) (C5). provider process 에 flag/env 전달은 C4 Provider Runtime 후속 migration slice 가 소유한다.

이 SSOT 는 **C7 Security / Policy** context 를 채운다. C7 은 cross-cutting 으로 C3·C4·C5·C6·C8 모두에 결정을 공급한다. Context map SSOT 는 [`./context-map.md`](./context-map.md) 가 소유한다.

## 0. 핵심 invariant (단단히 박는다)

다음 여섯 invariant 는 본 도메인 전체의 1차 약속이다. 본문 다른 절은 이들의 정교화다.

1. **`ExposesUnsafePermissionBypass=true` 는 risk signal 이며 사용 허가가 아니다.** capability 가 우회 surface 를 노출했다는 사실만으로 그것을 켤 수 없다 — 사용 가부는 본 문서의 policy bundle 게이트가 단독 결정한다. C3 Provider Capability SSOT 는 public `riido-contracts/provider/capability` 가 소유한다.

2. **`Host` trust tier 에서는 provider-native approval bypass 모드 활성화 절대 금지.** 다음 surface 가 모두 본 invariant 의 대상이다:
   - Claude `--permission-mode bypassPermissions`
   - Codex `--yolo`
   - Codex `--dangerously-bypass-approvals-and-sandbox`
   - wrapper 매니페스트가 자기신고한 동등 surface

   이 모드들은 **isolated trust tier 에서만**, **policy bundle 이 명시적으로 허용** 할 때만 사용 가능하다.

   Codex `--sandbox danger-full-access` 는 이 unsafe approval-bypass 목록이 아니다.
   뤼이도 daemon 은 provider 를 사용자 PC 에서 실제로 작업시키는 로컬 automation 으로
   취급하며, Codex 의 현재 canonical 실행 envelope 는 C4 adapter 가 명시적으로
   선택한 full-access runtime 이다. 이 선택은 provider / caller 의 default sandbox
   에 기대는 것이 아니라 “전권 실행을 인지한 harness-managed host runtime” 을 쓰는
   결정이다. caller custom arg / policy bundle 이 임의로 바꾸는 surface 가 아니다.

3. **default-deny.** policy bundle 이 명시적으로 허용하지 않은 모든 surface 는 자동으로 거절된다. 옛 정책 → 새 정책 으로 **downgrade(약화) 금지**. 정책 번들 버전은 항상 증가하는 방향으로만 적용된다.

4. **정책 결정과 실행의 분리.** 본 컨텍스트는 “이 surface 를 켜도 되는가?” / “이 path 가 보호 대상인가?” / “이 commit 이 destructive 한가?” 같은 **결정** 만 한다. 실제 flag 전달 / 파일 주입 / process 격리 / lint 실행 / lease 발급은 인접 context 가 수행한다.

5. **policy bundle 도 capability fingerprint 의 입력이다.** 정책 번들이 바뀌면 `CapabilityFingerprint` 가 바뀌고 lease 가 무효화된다. C3 Provider Capability SSOT 는 public `riido-contracts/provider/capability` 가 소유한다. 즉, **정책 변경은 silent 한 “더 허용해주기” 가 될 수 없다** — 진행 중 task 는 명시적으로 재평가된다.

6. **Store channel policy 는 보안 경계다.** `mac-app-store` / `msix-store` channel 에서 금지한 surface 는 provider capability 가 지원해도 사용할 수 없다. Provider CLI bundling, silent provider auto-install, 사용자 동의 없는 background helper, external TCP listener, arbitrary home scan 은 store channel 에서 항상 거절된다. Channel enum 과 role model 의 SSOT 는 [`./distribution-host-integration.md`](./distribution-host-integration.md) (C11) 이다.

## 1. Trust Tier

trust tier 는 “이 runtime 이 실행되는 환경의 격리 수준” 이다. 5 종으로 고정한다.

| Tier | 의미 | unsafe bypass 가능? |
| --- | --- | --- |
| `Host` | 데몬이 사용자/운영자 호스트 OS 위에 직접 실행 | **❌ 절대 금지** (§0 invariant 2) |
| `IsolatedContainer` | OS-level 컨테이너 격리 (Docker / containerd / Podman). 호스트 FS / network 가 격리됨 | policy bundle 명시 허용 시 |
| `EphemeralVM` | task 마다 1회용 VM (firecracker / micro-VM / 격리된 VM 인스턴스) | policy bundle 명시 허용 시 |
| `CIControlledRunner` | CI 시스템이 관리하는 격리 runner (예: GitHub Actions runner, GitLab runner) | policy bundle 명시 허용 시 — CI 의 추가 격리 보장이 있어야 함 |
| `Unknown` | trust tier 결정 불가 (미설정 / 매니페스트 부재 / 검증 실패) | **❌ 절대 금지** |

규칙:

1. trust tier 는 runtime 등록 시점에 **외부 신호** 로 결정된다. 자기 신고(wrapper 매니페스트, runtime 시작 환경 변수, host 정책 파일)와 검증 가능한 사실(예: cgroup / namespace 격리 detection)을 결합한다.
2. `Unknown` 으로 결정되면 capability 가 G-S1 (§3) 에서 거절된다.
3. trust tier 는 runtime 의 `ProviderCapability` 와 함께 `CapabilityFingerprint` 의 입력은 아니다(검증의 외부 사실). 단, 변경되면 capability 재평가가 트리거된다.

## 2. Policy Bundle

policy bundle 은 본 도메인의 **단일 결정 소스** 다. 모든 게이트는 활성 policy bundle 을 입력으로 받아 “허용/거절/보류” 를 답한다.

### 2.1 구조 (도메인 표현)

```
PolicyBundle {
    SchemaVersion    string             // "riido-policy-bundle.v1"
    Version           string             // opaque policy bundle version; 절대 downgrade 불가 (§0 invariant 3)
    EffectiveSince    time.Time
    SupersededAt      time.Time | null

    TrustTierPolicies map[TrustTier] TrustTierPolicy
    DefaultDeny       []PolicyTarget     // 명시적 black-list (예: 항상 거절할 path)
}

TrustTierPolicy {
    AllowedSurfaces   AllowedSurfaceSet   // 9 target 별 허용 표 (§3)
    Sandbox           SandboxPolicy
    NetworkEgress     EgressPolicy
    Secrets           SecretsPolicy
    ProtectedPaths    []PathPattern
    DestructiveOps    DestructiveOpPolicy
    MCPAllowlist      []MCPServerID
    NativeConfigRules NativeConfigPolicy
}
```

### 2.2 진화

- 새 번들은 **항상 새 Version**. 같은 Version 을 덮어쓰지 않는다.
- 두 번들의 차이가 “더 허용” 이라도 **각 task 는 시작 시점의 번들 버전** 으로 평가된다. silent expand 금지.
- 진행 중 task 가 새 번들로 “옮겨가려면” [`../30-architecture/runtime-upgrade-flow.md`](../30-architecture/runtime-upgrade-flow.md) 의 T-POLICY 흐름을 따른다.

### 2.3 actor

정책 번들은 **운영자(human)** 가 PR 로 배포한다. agent / daemon 이 정책을 변경하는 경로는 없다.

### 2.4 실행 artifact — `riido-policy-bundle.v1`

C7 policy bundle 의 현재 물리 형태는 **단일 JSON 파일** 이다. 파일 경로는 daemon 시작 환경의 `RIIDO_POLICY_BUNDLE_PATH` 로 주입하며, 파일 안의 `version` 이 활성 `PolicyBundleVersion` 이 된다. `RIIDO_POLICY_BUNDLE_VERSION` 을 함께 지정한 경우에는 파일 `version` 과 정확히 일치해야 하며, 다르면 daemon 설정 로드를 실패시킨다. 경로가 없으면 기존 local 개발 기본값 `policy-bundle.local.v0` 을 사용한다.

현재 executable loader 는 C7 의 최소 실행 부분인 `unsafe_bypass`, `native_config_hooks`, `native_config_files`, `tool_use` allowed surface 를 받는다. 알 수 없는 필드는 fail-closed 로 거절한다. `RIIDO_POLICY_BUNDLE_PATH` 가 없으면 daemon 은 built-in `policy-bundle.local.v0` 을 사용하며, 이 번들은 Host trust tier 에서 Claude audit-only command hook (`claude:command-hooks:audit`) 만 허용하고 unsafe bypass / native config file / tool-use risk surface 는 허용하지 않는다. Codex app-server 의 full-access sandbox envelope 는 native config file surface 가 아니라 C4 command builder 가 직접 선택하는 provider runtime launch shape 다. `RIIDO_POLICY_BUNDLE_VERSION` 만 지정한 dev-local 실행은 같은 built-in allowed surface 에 version tag 만 바꾼다.

```json
{
  "schema_version": "riido-policy-bundle.v1",
  "version": "policy-bundle.example.v1",
  "effective_since": "2026-05-27T00:00:00Z",
  "superseded_at": null,
  "trust_tier_policies": {
    "IsolatedContainer": {
      "allowed_surfaces": {
        "unsafe_bypass": [
          "codex:--yolo"
        ],
        "native_config_hooks": [
          "claude:command-hooks:audit"
        ],
        "native_config_files": [],
        "tool_use": [
          "tool:network-egress"
        ]
      }
    }
  }
}
```

검증 규칙:

1. `schema_version` 은 `riido-policy-bundle.v1` 이어야 한다.
2. `version`, `effective_since`, `trust_tier_policies` 는 필수다.
3. `superseded_at` 이 있으면 `effective_since` 보다 이후여야 한다.
4. `trust_tier_policies` 의 key 는 §1 의 trust tier enum 만 허용한다.
5. `allowed_surfaces.unsafe_bypass` 값은 §5 의 알려진 provider unsafe bypass surface 만 허용하며 중복될 수 없다.
6. `Host` / `Unknown` trust tier 는 bundle 이 어떤 값을 담아도 unsafe bypass surface 를 허용할 수 없다. loader 는 이 조합을 invalid bundle 로 거절한다.
7. `allowed_surfaces.native_config_hooks` 값은 본 문서의 알려진 provider hook surface 만 허용하며 중복될 수 없다. 현재 surface 는 `claude:command-hooks:audit` 뿐이다.
8. `Unknown` trust tier 는 native config hook surface 를 허용할 수 없다. runtime trust tier 가 확정되지 않으면 hook materialization 은 fail-closed 된다.
9. `allowed_surfaces.native_config_files` 값은 본 문서의 알려진 provider config file/home surface 만 허용하며 중복될 수 없다. `codex:config-home:task-scoped` 는 과거 호환을 위해 parser 가 아는 surface 로 남지만 현재 native config plan 은 이를 materialize 하지 않는다.
10. `Unknown` trust tier 는 native config file surface 를 허용할 수 없다. runtime trust tier 가 확정되지 않으면 provider config-home materialization 은 fail-closed 된다.
11. `allowed_surfaces.tool_use` 값은 본 문서 §6 의 알려진 tool-use risk surface 만 허용하며 중복될 수 없다. 현재 surface 는 `tool:network-egress`, `tool:protected-path-write`, `tool:secret-exposure`, `tool:destructive-command` 이다.
12. `Unknown` trust tier 는 tool-use risk surface 를 허용할 수 없다. runtime trust tier 가 확정되지 않으면 ToolUseSecurityGate 는 `interrupt-and-block` 으로 수렴한다.

## 3. Policy 대상 9 종

각 대상은 `AllowedSurfaceSet` 의 한 멤버이며 trust tier 별로 독립 결정된다.

| ID | Target | 정책 표현 |
| --- | --- | --- |
| T-PERM | permission mode | enum: 어떤 permission mode 가 허용되는가 (예: Claude `default`/`acceptEdits`/`plan`만 / `bypassPermissions` 거부) |
| T-SBX | sandbox mode | enum + provider 별 activation 결정 (`read-only` / `workspace-write` / policy-owned `danger-full-access`; Codex current full-access envelope 는 §4.3 처럼 C4-owned) |
| T-NET | network egress | 모드 + allowlist (default-off / explicit allowlist / unrestricted) |
| T-PATH | protected paths | path glob list (`.git/**`, `.env*`, secrets dirs, prod config 등) |
| T-SEC | secret exposure | scoped-token 정책 (TTL, scope 한정, env 전달 거부, 로그 redaction) |
| T-DESTR | destructive command | shell command 패턴 차단 (`rm -rf`, `dd of=/dev/`, `git push --force`, DB drop 등) |
| T-PUSH | git push / deploy / migration | repo, branch, env 별 허용 매트릭스 (예: `main` push 는 human approval 필요) |
| T-MCP | MCP server allowlist | 등록 가능한 MCP server id 목록 + transport 제한 |
| T-CFG | native config injection | task 별로 어떤 정책 파일이 workdir 에 주입되어야 하는가 (CLAUDE.md / AGENTS.md / hooks settings / wrapper manifest) |

각 target 의 enum 값과 의미는 본 문서가 소유한다. 인접 context 는 **결정 결과** 만 받는다(예: C4 는 “T-SBX → workspace-write” 라는 결과를 받아 provider 에 전달).

