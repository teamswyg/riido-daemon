// Package riidoapi exposes Riido's local-only daemon API.
//
// The API intentionally uses a tiny local JSON envelope: one local IPC request
// with a method and optional params, one JSON response envelope.
// It is the first surface that GUI/Zed integrations can consume without
// reading Riido's state files directly.
package riidoapi
