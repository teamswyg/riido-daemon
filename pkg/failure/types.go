package failure

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
