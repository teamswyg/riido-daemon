package failure

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func (e *Error) Is(target error) bool {
	if e == nil {
		return target == nil
	}
	return sentinelMatches(e.sentinel, target)
}

func (e *Error) Layer() Layer {
	if e == nil {
		return ""
	}
	return e.sentinel.Layer()
}

func (e *Error) Kind() Kind {
	if e == nil {
		return ""
	}
	return e.sentinel.Kind()
}

func (e *Error) Op() string {
	if e == nil {
		return ""
	}
	return e.op
}

func (e *Error) Message() string {
	if e == nil {
		return ""
	}
	return e.message
}
