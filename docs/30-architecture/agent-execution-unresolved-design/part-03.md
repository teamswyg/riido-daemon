# Agent Execution Unresolved Design: Part 03

[Back to agent-execution-unresolved-design.md](../agent-execution-unresolved-design.md)

## 8. RAG / Public-Private Guardrails

이 문서는 public-safe architecture fact 만 담는다. RAG index 에 넣을 수 있는 것은
vocabulary, ownership, non-secret rules, generated contract shape, 테스트 기준이다.

RAG index 에 넣으면 안 되는 것:

- private repository clone URL with credential
- raw token, bearer token, PAT, GitHub App installation token
- AWS account-specific live evidence payload
- signed URL, temporary credential, SSM parameter value
- production smoke response body containing customer/task private content

private repo auth 가 필요해지면 public `WorkspacePlan` 은 `auth_mode=token_ref` 와
opaque `auth_ref` 만 갖고, 실제 token exchange/IAM/secret rotation evidence 는
private infra/runtime evidence store 에 둔다. public/private RAG index 는 물리적으로
분리한다.

## 9. Open Decisions

| ID | Decision needed | Default until decided |
| --- | --- | --- |
| Q-EXEC-001 | private repo auth 를 GitHub App installation token, user PAT broker, 또는 org-level deploy key 중 무엇으로 할 것인가 | public repo only, private fail-closed |
| Q-EXEC-002 | long-lived provider process 를 provider별 default 로 둘 것인가, conversational task 에만 opt-in 할 것인가 | one-shot default, resume id 있으면 explicit resume |
| Q-EXEC-003 | web approval timeout 과 fallback terminal state 는 무엇인가 | `blocked` with `approval_timeout` |
| Q-EXEC-004 | `WorkspacePlan` 을 assignment DTO 에 직접 넣을지, linked execution plan id 로 둘지 | assignment snapshot inline minimal plan |
| Q-EXEC-005 | stream delta retention 을 얼마나 유지할 것인가 | final answer durable, deltas bounded/ephemeral |

## 10. Non-goals

- 이 설계는 private repo auth 를 바로 구현하지 않는다.
- public daemon 은 provider CLI binary 를 번들하거나 설치하지 않는다.
- daemon 은 client read-model UI copy 를 소유하지 않는다.
- workspace materializer 는 사용자의 원본 repo working tree 를 직접 mutate 하지 않는다.
- public repository 에 live deployment payload 또는 secret evidence 를 저장하지 않는다.
