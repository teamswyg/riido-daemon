//go:build windows

package securecredential

import (
	"syscall"
	"unsafe"
)

type windowsLockSecurity struct {
	owner      *syscall.SID
	descriptor unsafe.Pointer
}

type windowsLockACL struct {
	revision, reserved uint8
	size, count, pad   uint16
}

type windowsLockACE struct {
	kind, flags uint8
	size        uint16
	mask        uint32
}

func newWindowsLockSecurity(api *windowsLockAPI) (*windowsLockSecurity, error) {
	token, err := syscall.OpenCurrentProcessToken()
	if err != nil {
		return nil, ErrUnavailable
	}
	defer token.Close()
	user, err := token.GetTokenUser()
	if err != nil {
		return nil, ErrUnavailable
	}
	owner, err := user.User.Sid.Copy()
	if err != nil {
		return nil, ErrUnavailable
	}
	sid, err := owner.String()
	if err != nil {
		return nil, ErrUnavailable
	}
	text, err := syscall.UTF16PtrFromString("O:" + sid + "D:P(A;;FA;;;" + sid + ")")
	if err != nil {
		return nil, ErrUnavailable
	}
	s := &windowsLockSecurity{owner: owner}
	ok, _, _ := api.convertSD.Call(uintptr(unsafe.Pointer(text)), 1,
		uintptr(unsafe.Pointer(&s.descriptor)), 0)
	if ok == 0 || s.descriptor == nil {
		s.close()
		return nil, ErrUnavailable
	}
	return s, nil
}

func (s *windowsLockSecurity) close() {
	if s.descriptor != nil {
		_, _ = syscall.LocalFree(syscall.Handle(uintptr(s.descriptor)))
		s.descriptor = nil
	}
}

func (s *windowsLockSecurity) attributes() *syscall.SecurityAttributes {
	return &syscall.SecurityAttributes{
		Length:             uint32(unsafe.Sizeof(syscall.SecurityAttributes{})),
		SecurityDescriptor: uintptr(s.descriptor), InheritHandle: 0,
	}
}

func (a *windowsLockAPI) privateObject(handle syscall.Handle, expected *syscall.SID) bool {
	var owner *syscall.SID
	var dacl *windowsLockACL
	var descriptor unsafe.Pointer
	code, _, _ := a.securityInfo.Call(uintptr(handle), 1, 5,
		uintptr(unsafe.Pointer(&owner)), 0, uintptr(unsafe.Pointer(&dacl)), 0,
		uintptr(unsafe.Pointer(&descriptor)))
	if descriptor != nil {
		defer func() { _, _ = syscall.LocalFree(syscall.Handle(uintptr(descriptor))) }()
	}
	if code != 0 || descriptor == nil || owner == nil || dacl == nil {
		return false
	}
	if ok, _, _ := a.validSID.Call(uintptr(unsafe.Pointer(owner))); ok == 0 {
		return false
	}
	if equal, _, _ := a.equalSID.Call(uintptr(unsafe.Pointer(owner)), uintptr(unsafe.Pointer(expected))); equal == 0 {
		return false
	}
	var control uint16
	var revision uint32
	if ok, _, _ := a.descriptorControl.Call(uintptr(descriptor), uintptr(unsafe.Pointer(&control)), uintptr(unsafe.Pointer(&revision))); ok == 0 || control&0x1004 != 0x1004 {
		return false // SE_DACL_PRESENT | SE_DACL_PROTECTED
	}
	if valid, _, _ := a.validACL.Call(uintptr(unsafe.Pointer(dacl))); valid == 0 || dacl.count != 1 {
		return false
	}
	var ace *windowsLockACE
	if ok, _, _ := a.getACE.Call(uintptr(unsafe.Pointer(dacl)), 0, uintptr(unsafe.Pointer(&ace))); ok == 0 || ace == nil {
		return false
	}
	// One explicit FILE_ALL_ACCESS allow ACE for the process user only. Null,
	// inherited, object-specific, callback and unfamiliar ACLs stay closed.
	if ace.kind != 0 || ace.flags != 0 || ace.mask != 0x1f01ff || ace.size < 20 {
		return false
	}
	sid := unsafe.Add(unsafe.Pointer(ace), unsafe.Sizeof(*ace))
	if valid, _, _ := a.validSID.Call(uintptr(sid)); valid == 0 {
		return false
	}
	equal, _, _ := a.equalSID.Call(uintptr(sid), uintptr(unsafe.Pointer(expected)))
	return equal != 0
}
