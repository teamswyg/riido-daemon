# RIID-4630: ApprovalRequested Timeout Owner Cleanup

[Back to daemon-lifecycle-cli](../daemon-lifecycle-cli.md)

This slice closes the public daemon `Q-RT-003` open question by moving the
approval wait timeout decision into the provider-runtime SSOT.

This slice does:

- state that C4 session actor run clocks own approval wait timeout policy
- remove `Q-RT-003` from daemon open questions
- keep `EventIngestor` as an append authority only, not a timeout owner
- keep UI/review surfaces as display/response senders, not terminal timeout
  sources
- add a focused public workflow that fails if `Q-RT-003` drifts back into open
  questions
- add a reducer test that `EventToolApprovalNeeded` resets the semantic idle
  watchdog

This slice does not change provider-native approval RPC frames, add UI, change
CLI flags, introduce dependencies, or alter hard/semantic timeout defaults.
