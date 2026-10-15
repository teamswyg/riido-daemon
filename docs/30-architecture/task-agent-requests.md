# Agent 요청 전달 · RIID-5828

2026-10-12

댓글 멘션과 직접 배정 요청은 Control Plane에서 받은 assignment 실행 ID를 유지. 선택적 request_id와 workspace_id를 TaskRequest metadata로 전달하며, 사용자 추가 요청은 Prompt에 유지. system prompt나 Agent instruction으로 이동하지 않음.

검증: SaaS plane 요청 변환 테스트에서 요청별 실행 ID, 재시도 ID 유지, workspace/request metadata와 구버전 fallback 확인.

배포 선행 조건: riido-contracts 새 버전 및 Control Plane의 요청 계약 반영.
