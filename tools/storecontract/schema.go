package main

const (
	contractSchemaVersion = "riido-store-distribution-contract.v1"
	checkSchemaVersion    = "riido-store-distribution-contract-check.v1"
)

var storeReviewSubmissionRequiredSurfaces = []string{
	"demo-review-account",
	"privacy-metadata-allowlist",
	"provider-non-bundling-review-note",
}
