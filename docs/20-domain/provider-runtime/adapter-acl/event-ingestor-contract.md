# EventIngestor Contract

[Back to adapter-acl.md](../adapter-acl.md)

EventIngestor 의 append authority contract 는 public `riido-contracts/docs/20-domain/ir-event-log.md` §5.0 가 소유하고, daemon-side 구현은 public `internal/ir/ingest` 가 소유한다. C4 와 ingest 사이의 계약은 이 문서가 고정한다.

## RunController

`Provider.Drafts()` 채널을 읽어서 EventIngestor 의 single append API 를 호출하는 주체는 adapter 구현체가 아니라 **RunController** 다. RunController 는 C4 orchestration layer 다.

책임:

- 한 run 동안 `Provider.Drafts()` 채널을 drain.
- 받은 `ProviderEventDraft` 를 EventIngestor 의 단일 API 로 넘긴다. 현재 public Go API 는 `ingest.Ingestor.Append(ctx, ingest.Draft)` 다.
- adapter lifecycle 호출(`Cancel` / `Interrupt` / `ProvideInput` / `ResolveApproval`) 을 외부 orchestrator 신호에 따라 수행한다.
- adapter 가 `Drafts()` 채널을 닫으면 run lifecycle 을 종료시키고 cleanup 한다.

Adapter 구현체는 EventIngestor 를 모른다. RunController 가 `Provider.Drafts()` 를 drain 하고 EventIngestor single append API 를 authorized caller 로서 호출한다.

## Flow

```text
provider raw stdout/RPC
   -> adapter ACL conversion and first-pass C7 redaction
ProviderEventDraft
   -> Provider.Drafts()
RunController (C4 orchestration, authorized caller)
   -> EventIngestor single append API
EventIngestor
   -> identity / ordering / runtime identity / attribution / schema / timestamp policy
   -> second-pass C7 redaction / audit check
CanonicalEvent (append-only)
```

## Responsibility Table

| Direction | Input | Output | Actor |
| --- | --- | --- | --- |
| Adapter -> Drafts() | provider raw stdout/RPC | `ProviderEventDraft` | adapter implementation |
| Drafts() -> ingest | `Provider.Drafts()` channel drain | EventIngestor single append API call | **RunController (C4 orchestration)** |
| other authorized caller -> ingest | API / validation / scheduler / operator signal | EventIngestor single append API call | FSM orchestrator, server transition layer, validation runner, runtime scheduler |
| ingest -> append | draft + lease lookup + actor policy + active schema | identity / ordering / runtime identity / attribution / schema / timestamp + `CanonicalEvent` | EventIngestor |
| RunController -> adapter | cancel / interrupt / provideInput / resolveApproval | adapter lifecycle call | RunController |

## Rules

1. `CanonicalEvent` append has exactly one code path: EventIngestor single append API.
2. Adapter implementations do not import EventIngestor.
3. Reducer cannot call EventIngestor; it is pure.
4. RunController lives outside adapter implementations.
5. RunController also ingests non-draft signals such as provider process death or fatal stderr.
6. If ingest is down, RunController preserves the no-drop backpressure contract.
