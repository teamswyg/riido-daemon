package hostintegration

// ToolLoginStatus is intentionally not a failure enum. LoginRequired means the
// provider is real but not currently routable.
type ToolLoginStatus string

const (
	ToolLoginUnknown  ToolLoginStatus = "unknown"
	ToolLoginLoggedIn ToolLoginStatus = "logged-in"
	ToolLoginRequired ToolLoginStatus = "login-required"
)

// Valid reports whether status is one of the SSOT-defined login statuses.
func (s ToolLoginStatus) Valid() bool {
	switch s {
	case ToolLoginUnknown, ToolLoginLoggedIn, ToolLoginRequired:
		return true
	default:
		return false
	}
}
