# Raw To Draft Rules

[Back to adapter-acl.md](../adapter-acl.md)

본 문서가 강제하는 변환 규칙:

1. **알려진 raw type -> 도메인 `EventType` 매핑.** 매핑 표는 어댑터마다 자기 코드 안에 두지만 정규화된 `Type` 은 public `riido-contracts/docs/20-domain/ir-event-log.md` §3 카탈로그에 등록된 것만 사용한다.
2. **알려지지 않은 raw type** -> `Type=ProviderUnknownEvent`, `RawType=<원본>`, `Raw=<페이로드>`. FSM transition 절대 발생시키지 않는다.
3. **알려진 raw type 이지만 모르는 raw 필드** -> 알려진 필드는 정규화된 `Payload` 에, 모르는 필드는 `Unknown` 으로 보존. drop 금지.
4. **해석으로 의미가 추가된 경우** -> `Payload.derived=true` 를 표기한다. 예: provider 가 "파일 수정" 을 자연어로만 말한 것을 `FileChanged` 로 추론한 경우.
5. **provider 가 transition-after-side 사실을 보고** (`RunReportedDone` 등) -> adapter 는 draft 를 발행하지만 transition 자체는 ingest 가 결정한다.
