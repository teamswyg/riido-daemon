//go:build windows

package riidoapi

type namedPipeAddr string

func (a namedPipeAddr) Network() string { return string(LocalTransportWindowsNamedPipe) }
func (a namedPipeAddr) String() string  { return string(a) }
