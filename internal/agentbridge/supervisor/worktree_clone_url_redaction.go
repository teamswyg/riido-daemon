package supervisor

import (
	"net/url"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

func redactedRepositoryURL(parsed *url.URL) string {
	if parsed == nil {
		return ""
	}
	copyURL := *parsed
	copyURL.User = nil
	copyURL.RawQuery = ""
	copyURL.ForceQuery = false
	copyURL.Fragment = ""
	copyURL.RawFragment = ""
	copyURL.RawPath = ""
	if fullName := assignmentcontract.NormalizePublicGitHubRepositoryFullName(copyURL.Path); fullName != "" {
		copyURL.Path = "/" + fullName
	} else {
		copyURL.Path = "/redacted"
	}
	return copyURL.String()
}
