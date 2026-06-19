# Cursor Real CLI Gate

[Back to Integration Gates](../integration-gates.md)

A-54 부터 Cursor real CLI integration gate 는 `ResultCompleted` 와 함께 daemon 이
선택한 workdir 안의 expected file artifact 를 확인한다.

이 gate 는 `AGENTBRIDGE_INTEGRATION=1` 과 local Cursor Agent auth/runtime 이 준비된
operator environment 에서만 실행된다.

Cursor adapter 는 native `--workspace <cwd>` 와 `--trust` 를 사용해 daemon-selected
workdir 을 전달하며, 이 gate 는 `--yolo` 없이 파일 side-effect 를 확인해야 한다.

Gate 가 skip 된 경우에는 filesystem side-effect 가 검증된 것이 아니다.
