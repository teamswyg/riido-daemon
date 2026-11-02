//go:build windows

package securecredential

import (
	"context"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

const (
	windowsLockReadControl    = 0x20000
	windowsLockReadAttributes = 0x80
)

func windowsLockPath(path string) (string, []string, error) {
	path = filepath.FromSlash(path)
	volume := filepath.VolumeName(path)
	if len(volume) != 2 || volume[1] != ':' || (volume[0] < 'A' || volume[0] > 'Z') && (volume[0] < 'a' || volume[0] > 'z') || !filepath.IsAbs(path) {
		return "", nil, ErrUnavailable
	}
	parts := strings.Split(strings.TrimPrefix(path[len(volume):], `\`), `\`)
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || strings.ContainsAny(part, ":*?\"<>|\x00") || strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") || !filepath.IsLocal(part) {
			return "", nil, ErrUnavailable
		}
	}
	return volume + `\`, parts, nil
}

func openWindowsLockDirectories(ctx context.Context, api *windowsLockAPI, security *windowsLockSecurity, path string) ([]syscall.Handle, string, error) {
	root, parts, err := windowsLockPath(path)
	if err != nil {
		return nil, "", err
	}
	rootName, err := syscall.UTF16PtrFromString(root)
	if err != nil {
		return nil, "", ErrUnavailable
	}
	if kind, _, _ := api.driveType.Call(uintptr(unsafe.Pointer(rootName))); kind != 3 {
		return nil, "", ErrUnavailable // Local fixed disks only; no UNC/mapped share.
	}
	var handles []syscall.Handle
	current := root
	for index := -1; index < len(parts); index++ {
		if err := ctx.Err(); err != nil {
			closeWindowsLockHandles(handles)
			return nil, "", err
		}
		if index >= 0 {
			current = filepath.Join(current, parts[index])
		}
		var handle syscall.Handle
		if index < 0 {
			handle, err = openWindowsLockRoot(root)
		} else {
			handle, err = openWindowsLockChildDirectory(api, security, handles[len(handles)-1], parts[index])
		}
		if err != nil {
			closeWindowsLockHandles(handles)
			return nil, "", err
		}
		handles = append(handles, handle)
		if index == len(parts)-1 && !api.privateObject(handle, security.owner) {
			closeWindowsLockHandles(handles)
			return nil, "", ErrUnavailable
		}
	}
	return handles, current, nil
}

func openWindowsLockRoot(path string) (syscall.Handle, error) {
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return syscall.InvalidHandle, ErrUnavailable
	}
	// Denying delete sharing pins the volume root. Attribute-only writes are
	// not excluded by sharing; descendants use handle-relative no-reparse opens.
	handle, err := syscall.CreateFile(name, windowsLockReadAttributes|windowsLockReadControl,
		syscall.FILE_SHARE_READ, nil, syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS|syscall.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return syscall.InvalidHandle, ErrUnavailable
	}
	if !windowsLockFileKind(handle, true) {
		_ = syscall.CloseHandle(handle)
		return syscall.InvalidHandle, ErrUnavailable
	}
	return handle, nil
}

func openWindowsLockChildDirectory(api *windowsLockAPI, security *windowsLockSecurity, parent syscall.Handle, name string) (syscall.Handle, error) {
	if !windowsLockFileKind(parent, true) {
		return syscall.InvalidHandle, ErrUnavailable
	}
	handle, err := api.openRelative(security, parent, name, true)
	if err != nil {
		return syscall.InvalidHandle, ErrUnavailable
	}
	if !windowsLockFileKind(parent, true) || !windowsLockFileKind(handle, true) {
		_ = syscall.CloseHandle(handle)
		return syscall.InvalidHandle, ErrUnavailable
	}
	return handle, nil
}

func openWindowsLockFile(api *windowsLockAPI, security *windowsLockSecurity, parent syscall.Handle, name string) (syscall.Handle, error) {
	if !windowsLockFileKind(parent, true) || !api.privateObject(parent, security.owner) {
		return syscall.InvalidHandle, ErrUnavailable
	}
	handle, err := api.openRelative(security, parent, name, false)
	if err != nil {
		return syscall.InvalidHandle, ErrUnavailable
	}
	if !windowsLockFileKind(parent, true) || !api.privateObject(parent, security.owner) || !windowsLockFileKind(handle, false) || !api.privateObject(handle, security.owner) {
		_ = syscall.CloseHandle(handle)
		return syscall.InvalidHandle, ErrUnavailable
	}
	return handle, nil
}

func windowsLockFileKind(handle syscall.Handle, directory bool) bool {
	kind, err := syscall.GetFileType(handle)
	if err != nil || kind != syscall.FILE_TYPE_DISK {
		return false
	}
	var info syscall.ByHandleFileInformation
	if syscall.GetFileInformationByHandle(handle, &info) != nil || info.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return false
	}
	if (info.FileAttributes&syscall.FILE_ATTRIBUTE_DIRECTORY != 0) != directory {
		return false
	}
	return directory || info.NumberOfLinks == 1
}
