// Package processexec is the os/exec implementation of the process port.
//
// The implementation spawns a single child process via exec.CommandContext,
// fans out stdout/stderr/exit through bounded channels, and writes stdin
// through the command pipe behind a small synchronization guard.
package processexec
