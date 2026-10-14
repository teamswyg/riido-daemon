package main

import "testing"

type contractMutationCase struct {
	name   string
	mutate func(*contract)
	error  string
}

func expectContractMutationFailures(t *testing.T, cases []contractMutationCase) {
	t.Helper()
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			expectContractMutationFailure(t, tt.mutate, tt.error)
		})
	}
}

func expectContractMutationFailure(t *testing.T, mutate func(*contract), wanted string) {
	t.Helper()
	root := t.TempDir()
	writeRequiredDocs(t, root)
	value := validContract()
	mutate(&value)
	writeContract(t, root, value)

	result, err := run(root, "packaging/store/riido_daemon_store_distribution.riido.json")
	if err == nil {
		t.Fatalf("expected contract error containing %q", wanted)
	}
	if !hasError(result.Errors, wanted) {
		t.Fatalf("expected %q, got %v", wanted, result.Errors)
	}
}
