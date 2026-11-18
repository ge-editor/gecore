package remote

import (
	"encoding/json"
	"net"
	"os"
	"time"
)

// Handler executes a request inside ge. It is called from a connection
// goroutine, not from the UI goroutine: an implementation must hand the work
// over to the main loop (taking uiMu or posting an event) and return only
// when the operation has finished, because the CLI treats the response as
// the result of the operation.
type Handler interface {
	Handle(Request) Response
}

// HandlerFunc adapts a function to Handler.
type HandlerFunc func(Request) Response

func (f HandlerFunc) Handle(r Request) Response { return f(r) }

const (
	serverReadTimeout  = 5 * time.Second
	serverWriteTimeout = 10 * time.Second
)

// Server accepts remote requests on this process's socket.
type Server struct {
	l    net.Listener
	path string
	h    Handler
}

// Listen starts serving on the socket of process pid (normally os.Getpid()).
func Listen(pid int, h Handler) (*Server, error) {
	if err := os.MkdirAll(SocketDir(), 0o700); err != nil {
		return nil, err
	}
	path := SocketPath(pid)
	os.Remove(path) // left over from an earlier process that had this PID
	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	_ = os.Chmod(path, 0o600)
	s := &Server{l: l, path: path, h: h}
	go s.acceptLoop()
	return s, nil
}

// Close stops serving and removes the socket.
func (s *Server) Close() error {
	err := s.l.Close()
	os.Remove(s.path)
	return err
}

func (s *Server) acceptLoop() {
	for {
		c, err := s.l.Accept()
		if err != nil {
			return // listener closed
		}
		go s.serve(c)
	}
}

func (s *Server) serve(c net.Conn) {
	defer c.Close()
	var req Request
	c.SetReadDeadline(time.Now().Add(serverReadTimeout))
	if err := json.NewDecoder(c).Decode(&req); err != nil {
		return // liveness probe (connect and close) or garbage
	}
	c.SetReadDeadline(time.Time{})
	resp := s.h.Handle(req)
	c.SetWriteDeadline(time.Now().Add(serverWriteTimeout))
	_ = json.NewEncoder(c).Encode(resp)
}
