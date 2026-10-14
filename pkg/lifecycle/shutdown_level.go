package lifecycle

type ShutdownLevel uint8

const (
	ShutdownNone ShutdownLevel = iota
	ShutdownGraceful
	ShutdownForced
)

func (l ShutdownLevel) String() string {
	switch l {
	case ShutdownNone:
		return "none"
	case ShutdownGraceful:
		return "graceful"
	case ShutdownForced:
		return "forced"
	default:
		return "unknown"
	}
}

func (l ShutdownLevel) AtLeast(want ShutdownLevel) bool {
	return l >= want
}

func (l ShutdownLevel) IsShutdown() bool {
	return l.AtLeast(ShutdownGraceful)
}

func (l ShutdownLevel) IsForced() bool {
	return l.AtLeast(ShutdownForced)
}
