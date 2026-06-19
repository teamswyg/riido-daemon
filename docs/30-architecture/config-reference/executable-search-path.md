# Executable Search Path

[Back to Daemon Config Reference](../config-reference.md)

When no `RIIDO_<PROVIDER>_PATH` override is set, `internal/agentbridge/detectutil`
uses an augmented search path. This protects Desktop/launchd/service daemons from
minimal inherited PATH values such as `/usr/bin:/bin:/usr/sbin:/sbin`.

Search order:

1. process `PATH`; explicit operator PATH still wins.
2. user login-shell PATH via `$SHELL -lc`, cached once per process and skipped on
   Windows, unset shell, or timeout.
3. known install dirs: `/opt/homebrew/{bin,sbin}`, `/usr/local/{bin,sbin}`,
   standard system dirs, `~/.local/bin`, `~/.npm-global/bin`, `~/.cargo/bin`,
   `~/.bun/bin`, `~/.deno/bin`, `~/go/bin`, `~/.volta/bin`, `~/.asdf/shims`,
   `~/.cursor/bin`, `~/.claude/bin`, plus resolved nvm/fnm/asdf Node bins.

This only widens unset-override lookup. Explicit `RIIDO_<PROVIDER>_PATH` still
resolves exactly that file and never falls back.
