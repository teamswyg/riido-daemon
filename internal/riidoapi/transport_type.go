package riidoapi

type LocalTransport string

const (
	LocalTransportUnixSocket       LocalTransport = "unix-socket"
	LocalTransportWindowsNamedPipe LocalTransport = "windows-named-pipe"
)
