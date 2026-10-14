package codex

func toolchainPermissionProfileTokens() []string {
	return []string{
		`"/usr/local/go"="read"`,
		`"/Users/example/.rustup"="read"`,
		`"/Users/example/.cargo"="read"`,
		`"/Users/example/Library/Caches/go-build"="write"`,
		"default_permissions",
	}
}
