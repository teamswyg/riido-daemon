# 6-7 NativeConfigVersion Rules

[Back to native config manifest](../native-config-manifest.md)

`NativeConfigVersion` 은 execution-bound `CanonicalEvent` 의 의무 필드다. 본
문서가 생성 규칙을 소유하고, event schema 자체는 public `riido-contracts` C2
계약이 소유한다.

```text
NativeConfigVersion = sha256-hex(
   canonicalJSON({
       policyBundleVersion: <C7 활성 번들 버전>,
       nativeConfigPlan: {
           providerKind:       ...,
           protocolKind:       ...,
           injectedFiles[]:    [{ path, sha256(content) }, ...],
           hookScriptVersions: [{ id, sha256 }, ...],
           wrapperManifestSha: <opt> ,
       },
       schemaVersion: 1
   })
)
```

규칙:

1. 입력에 **모든 주입된 파일의 내용 해시** 가 포함되어야 한다 → 한 줄만 바뀌어도 새 버전.
2. 입력에 `policyBundleVersion` 이 포함되어야 한다 → 정책 번들 변경은 항상 NativeConfigVersion 변경.
3. 주입된 파일에는 `.riido/native-config-manifest.json` 도 포함된다. provider filename catalog, hook materialization mode, telemetry placement 의 변경은 manifest 내용 변경으로 NCV 에 반영된다.
4. 알고리즘 / 입력 schema 자체가 바뀌면 `schemaVersion` 을 올린다. 옛 schemaVersion 의 산출값은 영구 보존 (replay 호환).
5. `NativeConfigVersion` 은 task 시작 시점에 정해진 뒤 그 run 동안 **불변**. 변경하려면 `ReinjectNativeConfig` 또는 새 run (`ReworkQueued → Queued`).
6. local daemon 의 supervisor 는 native config 주입 직후 이 값을 계산해 run metadata `native_config_version` 에 고정한다. `NativeConfigInjected` / `WorkdirArchived` 같은 Cat E 이벤트 append 는 같은 `NativeConfigVersion` 을 EventIngestor 경로로 stamp 한다.

## 7. PolicyBundleVersion ↔ NativeConfigVersion 관계

- `PolicyBundleVersion` 변경 → `NativeConfigVersion` 변경 (§6 입력의 한 멤버이므로).
- `NativeConfigVersion` 변경이 `PolicyBundleVersion` 변경을 함의하지는 않는다 (예: 같은 정책 번들 + 새 wrapper manifest).
- 둘 다 task 시작 시점에 함께 고정 → execution-bound CanonicalEvent 필드로 영속화.
- 진행 중 task 가 두 값 중 하나라도 silent 하게 따라가는 것을 금지. 변경은 runtime upgrade flow 의 T-POLICY / T-CONFIG 분기를 통해서만 가능하며, 해당 architecture SSOT 는 provider-runtime slice 에서 public repo 로 이동한다.
