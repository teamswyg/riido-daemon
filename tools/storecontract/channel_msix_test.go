package main

import "testing"

func TestRunRejectsMSIXSideloadWithoutSignedPackage(t *testing.T) {
	expectContractMutationFailure(
		t,
		removeChannelRequiredSurface("msix-sideload", "signed-msix-package"),
		`channel "msix-sideload" must require signed-msix-package`,
	)
}

func TestRunRejectsMSIXStoreWithoutNamedPipeIPC(t *testing.T) {
	expectContractMutationFailure(
		t,
		removeChannelRequiredSurface("msix-store", "named-pipe-local-ipc"),
		`channel "msix-store" must require named-pipe-local-ipc`,
	)
}
