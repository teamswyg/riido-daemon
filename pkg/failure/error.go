package failure

type Error struct {
	sentinel Sentinel
	op       string
	message  string
	cause    error
}

func New(sentinel Sentinel, op, message string) error {
	return &Error{sentinel: sentinel, op: op, message: message}
}

func Wrap(sentinel Sentinel, op, message string, cause error) error {
	if cause == nil {
		return New(sentinel, op, message)
	}
	return &Error{sentinel: sentinel, op: op, message: message, cause: cause}
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.errorString()
}

func (e *Error) errorString() string {
	prefix := e.sentinel.Error()
	if e.op != "" {
		prefix += ": " + e.op
	}
	if e.message != "" {
		prefix += ": " + e.message
	}
	if e.cause != nil {
		prefix += ": " + e.cause.Error()
	}
	return prefix
}
