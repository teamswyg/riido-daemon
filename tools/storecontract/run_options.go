package main

const defaultPolicyTablePath = "docs/20-domain/distribution-host-integration/store-channel-policy/policy-table.md"

type runOptions struct {
	PolicyTablePath  string
	WritePolicyTable bool
	CheckPolicyTable bool
}

func (opts runOptions) policyTableEnabled() bool {
	return opts.WritePolicyTable || opts.CheckPolicyTable
}
