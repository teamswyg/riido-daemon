# Local Daemon Flags

[Back to Daemon Config Reference](../config-reference.md)

| Flag | Default | Meaning |
| --- | --- | --- |
| `--socket` | C11 dev-local local IPC endpoint | local API socket or named pipe path |
| `--transport` | `unix-socket` | `unix-socket` or `windows-named-pipe` |
| `--pid-file` | unset | optional background PID file |
| `--log-file` | unset | optional structured log file |
| `--foreground` | false | run daemon in the current process |
| `--lock-file` | `$HOME/.riido/.lock` | local singleton advisory lock |
| `--timeout-seconds` | `5` for stop | graceful stop wait before kill |
| `--force` | false | request forced shutdown before PID fallback |

`cmd/riido` remains local-only. Health, ready, status, and metrics are exposed
through local IPC subcommands, not public HTTP.
