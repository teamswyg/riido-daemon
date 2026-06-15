package failure

import (
	"errors"
	"fmt"
)

type Layer string

type Kind string

type Layered interface {
	Layer() Layer
}

type Classified interface {
	Layered
	Kind() Kind
}

type Operational interface {
	Op() string
}

type Messaged interface {
	Message() string
}

type Sentinel struct {
	layer Layer
	kind  Kind
}

func NewSentinel(layer Layer, kind Kind) Sentinel {
	return Sentinel{layer: layer, kind: kind}
}

func (s Sentinel) Error() string {
	if s.layer == "" {
		return string(s.kind)
	}
	if s.kind == "" {
		return string(s.layer)
	}
	return string(s.layer) + "/" + string(s.kind)
}

func (s Sentinel) Layer() Layer {
	return s.layer
}

func (s Sentinel) Kind() Kind {
	return s.kind
}

func (s Sentinel) Is(target error) bool {
	return sentinelMatches(s, target)
}

type Error struct {
	sentinel Sentinel
	op       string
	message  string
	cause    error
}

func New(sentinel Sentinel, op string, message string) error {
	return &Error{sentinel: sentinel, op: op, message: message}
}

func Wrap(sentinel Sentinel, op string, message string, cause error) error {
	if cause == nil {
		return New(sentinel, op, message)
	}
	return &Error{sentinel: sentinel, op: op, message: message, cause: cause}
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
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

func Is(err error, sentinel Sentinel) bool {
	return errors.Is(err, sentinel)
}

func AsClassified(err error) (Classified, bool) {
	var classified Classified
	if errors.As(err, &classified) {
		return classified, true
	}
	return nil, false
}

func AsOperational(err error) (Operational, bool) {
	var operational Operational
	if errors.As(err, &operational) {
		return operational, true
	}
	return nil, false
}

func sentinelMatches(sentinel Sentinel, target error) bool {
	if target == nil {
		return false
	}
	switch t := target.(type) {
	case Sentinel:
		return sentinel == t
	case *Sentinel:
		return t != nil && sentinel == *t
	case Classified:
		return sentinel.Layer() == t.Layer() && sentinel.Kind() == t.Kind()
	default:
		return false
	}
}

func Format(message string, args ...any) string {
	if len(args) == 0 {
		return message
	}
	return fmt.Sprintf(message, args...)
}
