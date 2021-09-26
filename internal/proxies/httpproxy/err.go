package httpproxy

var (
	errShutdown = proxyErr{Err: "server shutdown", temp: false}

	// used when we hijack for CONNECT
	_200Ok []byte = []byte("HTTP/1.0 200 OK\r\n\r\n")
)

type proxyErr struct {
	error
	Err  string
	temp bool
}

// net.Error interface implementation
func (e *proxyErr) String() string  { return e.Err }
func (e *proxyErr) Temporary() bool { return e.temp }
func (e *proxyErr) Timeout() bool   { return false }
