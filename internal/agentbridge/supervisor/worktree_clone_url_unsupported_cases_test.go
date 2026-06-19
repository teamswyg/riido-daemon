package supervisor

func unsupportedCloneURLCases() []unsupportedCloneURLCase {
	return []unsupportedCloneURLCase{
		{
			name:       "userinfo",
			repoURL:    "https://token:secret@example.com/teamswyg/riido-daemon",
			notContain: []string{"secret", "token"},
		},
		{
			name:       "query token",
			repoURL:    "https://github.com/teamswyg/riido-daemon?token=secret",
			notContain: []string{"secret", "token="},
		},
		{name: "empty force query", repoURL: "https://github.com/teamswyg/riido-daemon?"},
		{
			name:       "fragment token",
			repoURL:    "https://github.com/teamswyg/riido-daemon#secret-token",
			notContain: []string{"secret-token"},
		},
		{name: "missing repo path", repoURL: "https://github.com"},
		{
			name:       "full name query",
			fullName:   "teamswyg/riido-daemon?token=secret",
			notContain: []string{"secret", "token="},
		},
		{
			name:       "full name encoded query",
			fullName:   "teamswyg/riido-daemon%3Ftoken=secret",
			notContain: []string{"secret", "token="},
		},
	}
}
