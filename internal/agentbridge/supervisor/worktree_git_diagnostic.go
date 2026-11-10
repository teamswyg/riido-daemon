package supervisor

import "strings"

// Only fixed text is retained in this error. Git/helper output may contain
// credentials and must never be wrapped, logged, or sent to the control plane.
type assignmentGitCloneFailure struct {
	reason string
	action string
}

func (e *assignmentGitCloneFailure) Error() string {
	return "Repository preparation failed. Reason: " + e.reason + " Action: " + e.action
}

func classifyAssignmentGitCloneFailure(stderr []byte) *assignmentGitCloneFailure {
	output := strings.ToLower(string(stderr))
	const verifyAccess = "In a terminal on the daemon device, using the same OS account as Riido, run `git ls-remote <repository HTTPS URL>` with the repository's URL, then retry the task after it succeeds."
	const signIn = "Sign in to GitHub through this device's Git credential manager, or, if GitHub CLI is installed, run `gh auth login --hostname github.com --git-protocol https` followed by `gh auth setup-git --hostname github.com`. "
	switch {
	case strings.Contains(output, "remote branch") && strings.Contains(output, "not found"),
		strings.Contains(output, "couldn't find remote ref"):
		return &assignmentGitCloneFailure{
			reason: "The selected branch was not found in the remote repository.",
			action: "Select an existing branch in the task's repository settings, or push the intended branch to GitHub, then retry the task.",
		}
	case containsGitDiagnostic(output, "ssl certificate problem", "server certificate verification failed", "certificate verify failed", "unable to get local issuer certificate", "ssl peer certificate"):
		return &assignmentGitCloneFailure{
			reason: "Git could not verify the HTTPS certificate.",
			action: "Check the device clock and Git's trusted certificate chain. If a company proxy is used, ask its administrator to configure the required CA certificate for Git. " + verifyAccess,
		}
	case containsGitDiagnostic(output, "could not resolve host", "could not resolve proxy", "failed to connect", "couldn't connect", "connection timed out", "operation timed out", "connection reset", "network is unreachable"):
		return &assignmentGitCloneFailure{
			reason: "Git could not connect to the repository host or proxy.",
			action: "Check this device's network, DNS, VPN, proxy and firewall settings. " + verifyAccess,
		}
	case containsGitDiagnostic(output, "repository not found", "returned error: 404", "returned error: 403", "saml sso", "permission to "):
		return &assignmentGitCloneFailure{
			reason: "The repository is unavailable to this device's Git account. The URL may be incorrect, the repository may no longer exist, or the account may lack access.",
			action: "Verify the repository URL and sign in with an account that has read access. Ask the repository owner for access and authorize organization SSO if required. " + verifyAccess,
		}
	case containsGitDiagnostic(output, "authentication failed", "invalid username or password", "invalid username or token", "http basic: access denied", "returned error: 401"):
		return &assignmentGitCloneFailure{
			reason: "Git authentication was rejected. Saved credentials may be expired, revoked or for the wrong account.",
			action: "Refresh the saved Git credentials for an account with repository access. " + signIn + verifyAccess,
		}
	case containsGitDiagnostic(output, "could not read username", "could not read password", "terminal prompts disabled", "user interactivity has been disabled", "cannot prompt because", "interactive prompts have been disabled"):
		return &assignmentGitCloneFailure{
			reason: "Git authentication is required on this device, but the background daemon cannot open a login prompt. Riido's GitHub connection does not sign in Git on this device.",
			action: signIn + verifyAccess,
		}
	default:
		return &assignmentGitCloneFailure{
			reason: "Git could not prepare the repository; the exact cause could not be determined.",
			action: verifyAccess + " If access succeeds, check the selected branch and the daemon work directory's write permissions and free disk space. The terminal's local Git diagnostic can help identify the remaining cause.",
		}
	}
}

func containsGitDiagnostic(output string, patterns ...string) bool {
	for _, pattern := range patterns {
		if strings.Contains(output, pattern) {
			return true
		}
	}
	return false
}
