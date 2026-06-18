package detectutil

import "os"

// EnvOverride reads an env var by key, defaulting to os.Getenv.
// Provided for test injection — tests pass their own getter.
type EnvOverride func(key string) string

// OSEnv is the production env getter.
func OSEnv(key string) string { return os.Getenv(key) }
