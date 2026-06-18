package toolpolicy

func protectedFileNames() []string {
	return []string{
		".bash_profile", ".bashrc", ".claude.json", ".gitconfig",
		".gitmodules", ".mcp.json", ".netrc", ".npmrc", ".profile",
		".pypirc", ".ripgreprc", ".zprofile", ".zshrc", "credentials",
	}
}

func protectedDirectoryNames() []string {
	return []string{".git", ".vscode", ".idea", ".husky", ".claude", ".aws", ".ssh", ".gnupg"}
}
