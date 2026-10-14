package runtimeactor

import "errors"

var (
	// ErrUnknownProvider — Submit referenced a provider name the Actor
	// does not have an Adapter for.
	ErrUnknownProvider = errors.New("runtimeactor: unknown provider")
	// ErrSlotExhausted — MaxConcurrent reached. Policy: reject (the
	// caller may retry or queue externally). audit M-1 §HonorsMaxConcurrentSlots.
	ErrSlotExhausted = errors.New("runtimeactor: max concurrent reached")
	// ErrUnknownTask — Cancel referenced a taskID that is not in-flight.
	ErrUnknownTask = errors.New("runtimeactor: unknown task")
	// ErrActorStopped — Submit or Cancel after Stop.
	ErrActorStopped = errors.New("runtimeactor: stopped")
	// ErrProviderUnavailable — the provider's Detect reported Available=false.
	ErrProviderUnavailable = errors.New("runtimeactor: provider unavailable")
	// ErrDuplicateTaskID — Submit with a taskID that is already running.
	ErrDuplicateTaskID = errors.New("runtimeactor: duplicate task id")
	// ErrRuntimePinViolated — a running task's runtime identity changed.
	ErrRuntimePinViolated = errors.New("runtimeactor: runtime pin violated")
)
