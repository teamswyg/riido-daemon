# Change Procedure

[Back to Daemon Config Reference](../config-reference.md)

When an env var or daemon flag is added:

1. update this config reference in the same PR that reads it.
2. add or update parser tests.
3. keep failure modes explicit.
4. route provider-specific env reads through testable env helpers rather than
   direct unscoped `os.Getenv` calls in adapter internals.
