# Change Procedure

[Back to provider integration matrix](../integration-matrix.md)

When a provider adapter changes real CLI behavior, update the provider test and
this matrix in the same PR. New providers must add deterministic public tests,
an instruction placement strategy, and an effectiveness probe marker before
adding optional real-CLI integration.
