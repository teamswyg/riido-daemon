package hostintegration

// WorkspaceGrantMethod records the OS/user action that makes a user workspace
// root accessible to Riido.
type WorkspaceGrantMethod string

const (
	WorkspaceGrantDevLocalPath             WorkspaceGrantMethod = "dev-local-path"
	WorkspaceGrantUserSelectedFolder       WorkspaceGrantMethod = "user-selected-folder"
	WorkspaceGrantSecurityScopedBookmark   WorkspaceGrantMethod = "security-scoped-bookmark"
	WorkspaceGrantWindowsFolderPickerGrant WorkspaceGrantMethod = "windows-folder-picker-grant"
)

// Valid reports whether method is one of the SSOT-defined workspace grant
// methods.
func (method WorkspaceGrantMethod) Valid() bool {
	switch method {
	case WorkspaceGrantDevLocalPath,
		WorkspaceGrantUserSelectedFolder,
		WorkspaceGrantSecurityScopedBookmark,
		WorkspaceGrantWindowsFolderPickerGrant:
		return true
	default:
		return false
	}
}
