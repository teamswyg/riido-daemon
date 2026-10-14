// Package policy implements the C7 Security / Policy decision helpers.
//
// It owns small, pure policy decisions that adjacent contexts consult before
// turning a risk surface into concrete provider flags. It does not spawn
// provider processes, mutate task state, or inspect provider capabilities; C4
// / C5 / C6 execute the decisions this package returns.
package policy
