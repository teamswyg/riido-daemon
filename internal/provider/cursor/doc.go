// Package cursor owns the C4 run-scope adapter for the Cursor Agent CLI.
//
// As of 2026-05 the cursor-agent CLI exposes `-p`, `--output-format`,
// `--yolo`, `--workspace`, and `--trust` at the root level. The historical
// `chat` subcommand is no longer accepted on current builds.
package cursor
