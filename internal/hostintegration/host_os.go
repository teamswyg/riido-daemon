package hostintegration

// HostOS is the operating-system family that owns app data path semantics.
type HostOS string

const (
	HostOSDarwin  HostOS = "darwin"
	HostOSWindows HostOS = "windows"
)

// Valid reports whether os is one of the SSOT-defined host OS values.
func (os HostOS) Valid() bool {
	switch os {
	case HostOSDarwin, HostOSWindows:
		return true
	default:
		return false
	}
}
