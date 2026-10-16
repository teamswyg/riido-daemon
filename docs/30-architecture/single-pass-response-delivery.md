# Single-pass response delivery

This change retains one provider turn on the normal path. There is no classifier,
reviewer, summarizer or automatic quality-retry model call. Existing explicitly
configured verification continuations and provider recovery are unchanged.

## Contract

- Answer-only requests need no repository, file mutation or shell check.
- The latest user request controls scope; historical task instructions are context.
- Requested execution must be reported with observed evidence, not proposed steps.
- A provider success with no non-whitespace answer and no verified host result is
  a failed empty response, never a generic successful completion.
- Terminal acknowledgements use a deterministic key scoped to execution, lease
  and result, including retries beyond the HTTP retry budget.
- Empty history entries do not consume the recent-message budget.
- Buffer streamed text before network state updates. Failed delivery must not erase
  the only answer; final delivery can recover the accumulated text.
- Measure daemon claim-to-first-text, first delivery and terminal acknowledgement
  with its own clock. Measure server event handling and assignment age with the
  server clock. Do not subtract clocks across machines or call these UI paint time.
- Logs contain identifiers, status and timing only, never prompts, tokens or answers.

## Active-run scheduling

Long-poll only while there are no in-flight runs and no queued supervisor
messages. Preparing, running and pending terminal delivery all count as in-flight.
A claim earlier in the same runtime wave disables long-poll for later runtimes.
Otherwise a provider that has already buffered its events can lose the cancellation
wakeup and each timer tick can delay another event by an entire poll timeout.
Short polls remain enabled so other providers can accept work during a run.

## Verification and boundaries

Focused regressions cover normal single-turn answers, text-only completion,
whitespace/empty completion, failed first delivery and terminal fallback, scoped
history and the latest directive. Local authenticated Codex runs measure first SSE
answer and terminal state before/after with the same prompt and runtime-default model configuration. They are
small observations, not a statistical speedup claim. Test timing overhead separately
with local benchmarks; no model calls occur in ordinary unit tests.

No frontend, dependency, deployment, container isolation or new worktree lifecycle
is introduced. Existing session recovery and opt-in host checks remain authoritative.

## Local measurements — 2026-10-16

Answer-only task: summarize the supplied note in two Korean sentences, with no
repository, tools or file creation. macOS arm64 / Apple M3 Pro, Go 1.26.5,
Codex CLI 0.161.0, same runtime-default model configuration (resolved model ID
not pinned). Baselines: daemon `96304799a998791e2485d974a4a774c9166743cb`,
control-plane `82787a78f678300ce7dc9faafde4bca190ba56b9`.
Each run used disposable local server state, a separate local test account/device,
and the existing user's authenticated Codex CLI. No production state was used.

| Run | Before first SSE (ms) | After first SSE (ms) | Before terminal observed (ms) | After terminal observed (ms) |
| --- | ---: | ---: | ---: | ---: |
| 1 | 6324 | 7139 | 7051 | 8281 |
| 2 | 6751 | 12265 | 7255 | 12767 |
| 3 | 6824 | 6857 | 7300 | 7459 |
| 4 | 7494 | 7347 | 8138 | 8075 |
| 5 | 6275 | 6253 | 7064 | 6923 |
| Median | 6751 | 7139 | 7255 | 8075 |

Observed medians increased by 388 ms (5.7%) and 820 ms (11.3%). This is **not a
speedup or no-regression claim**: five unrandomized runs per revision, variable
provider latency and workstation load, one 12.265 s first-answer outlier, and no
pinned model ID. Runs 1–4 exercised the same answer path before the final terminal
retry-key/history-empty-row hardening; run 5 used the final functional code.
Provider/internal API request counts are not measured by session counts.

After runs recorded first-text-to-server-ack delays of 22, 9, 10, 16 and 11 ms.
For the outlier, claim-to-first-text was 11295 ms, then first delivery took 9 ms;
the delay occurred before answer text reached the daemon, not in its subsequent
HTTP delivery. This does not identify a particular upstream cause.
Timers start immediately before assignment POST after server/daemon readiness
and SSE connection. They include assignment scheduling; they exclude initial
service startup. Terminal observation polls every 100 ms. Baseline runs 1–3
used 10 ms SSE timestamp resolution. SSE receipt is not desktop UI paint time.
All ten runs produced streamed and final Korean answers; no manual desktop UI
validation is claimed here.

The opt-in native integration probe counted **one `turn/start` and one process
start**, with provider elapsed time 5272 ms. It checks no host verification was
required. This verifies the ordinary daemon turn path, not provider-internal
inference requests. Existing explicitly enabled continuations/recovery can still
start additional turns; this change adds none.

Prompt-construction microbenchmark (three samples, same 4800-character context):
72901/82149/93555 ns/op before, 20101/19714/17469 ns/op after; approximately
47176 B/op and 20 allocs/op in both. Both are below 0.1 ms in these samples;
these are prompt-building costs only, not model or end-to-end latency.

## Reproduce locally

Clone both PR branches into sibling `riido-daemon` and `riido-control-plane`
directories. From riido-daemon:

```sh
# riido-daemon
 go build -o /tmp/riido-response-daemon ./cmd/riido
```

Build `/tmp/riido-response-server` using the [control-plane build and verification instructions](https://github.com/teamswyg/riido-control-plane/blob/1d8280d816c6b10713d281b3ef2466d0e13a8672/docs/30-architecture/single-pass-response-delivery.md#reproduce-locally). Then, from riido-daemon:

```sh
python3 scripts/measure-single-pass-response.py \
  --daemon /tmp/riido-response-daemon \
  --server /tmp/riido-response-server
AGENTBRIDGE_INTEGRATION=1 go test ./internal/agentbridge/supervisor \
  -run '^TestIntegrationAnswerOnlyUsesOneProviderTurn$' -v -count=1
```

Requires a locally installed and authenticated Codex CLI; these two commands
consume actual provider usage. To pin a supported model, pass `--model MODEL_ID`
and set `AGENTBRIDGE_CODEX_MODEL=MODEL_ID` for the native test. The script prints
its measurements and private temporary evidence path and stops its own processes.
Do not commit the retained state: it includes disposable device credentials and
local run content. Existing desktop login/profile is not needed for this probe.

Ordinary validation (does not opt into real provider calls):

```sh
go test ./...
golangci-lint run ./... --timeout=5m
# riido-daemon
 go test -race ./internal/agentbridge/controlplane/saasplane ./internal/agentbridge/supervisor
```

Both full Go suites and the listed race checks passed. Lint is recorded in the PR.
Raw local logs are intentionally excluded; sanitized measurement values are kept
in the companion executable manifest.

## Loading regression — 2026-10-16

A local staging session showed provider event processing separated by repeated
25-second pauses while another runtime had no work. Pending events had already
cancelled an earlier claim, so a subsequent idle long-poll could still block the
single supervisor mailbox. Restrict long-poll to completely idle claim waves.

The deterministic active-run/queued-message probe exhausted its 200 ms deadline
before the change; afterwards the two claim waves took 38/17 microseconds in one
local observation. These values measure scheduling only, not model speed.
The original task page was verified with a one-line Claude answer after installing
an explicitly labelled local daemon build. Claim-to-first-text was 14425 ms,
first delivery acknowledgement 14844 ms, and terminal acknowledgement 18145 ms.
The answer and stopped loading indicator were verified in the browser. Initial
idle assignment discovery and provider time remain outside the fixed queue stall.

The client separately bounds an open-but-silent task SSE connection to 30 seconds,
then re-reads state and reconnects. This recovers missed terminal events; it neither
retries the task nor calls the model. This daemon change adds no model calls.
