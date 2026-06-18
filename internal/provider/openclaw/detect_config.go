package openclaw

const EnvOverride = "RIIDO_OPENCLAW_PATH"

// MinSupportedVersion is calendar-versioned. The parser only accepts 20XX
// years, so dependency errors with Node semver are never treated as OpenClaw.
const MinSupportedVersion = "2026.5.5"
