package claude

func safeStartOptions() StartOptions {
	return StartOptions{PermissionMode: PermissionModeApproval}
}
