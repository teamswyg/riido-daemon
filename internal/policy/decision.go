package policy

// Decision is the stable result shape returned by policy helpers.
type Decision struct {
	Allowed bool
	Code    string
	Reason  string
}
