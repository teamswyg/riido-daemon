# Responsibility Statement

[Back to runtime-responsibility.md](../runtime-responsibility.md)

Provider Runtime 은 **provider 의 실행 표면** (process, session, run, raw event stream)을
도메인 안으로 끌어들여 **정규화된 draft** 로 변환한다. 그것이 본 컨텍스트의 시작이고 끝이다.

C4 는 provider process/run lifecycle 과 adapter ACL 출력만 소유한다. It consumes an
already-approved runtime assignment input and produces provider-neutral runtime observations.
