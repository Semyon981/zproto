package zproto

import (
	"io"
	"log"
	"net"

	"github.com/Semyon981/zproto/zio"
	"github.com/Semyon981/zproto/zmux"
	"github.com/hashicorp/yamux"
)

type Handler interface {
	Handle(rw zio.ReadWriter)
}

type HandlerFunc func(rw io.ReadWriter)

func (f HandlerFunc) Handle(rw io.ReadWriter) {
	f(rw)
}

type Server struct {
	Handler Handler
}

func NewServer(Handler Handler) Server {
	return Server{Handler: Handler}
}

func (s *Server) Serve(l net.Listener) error {
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("failed to accept conn: %s\n", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	session, err := yamux.Server(conn, nil)
	if err != nil {
		log.Printf("failed to create session: %s\n", err)
		return
	}
	for {
		ch, err := session.Accept()
		if err != nil {
			log.Printf("failed to accept channel: %s\n", err)
			continue
		}
		go s.handleChannel(ch)
	}
}

func (s *Server) handleChannel(c net.Conn) {
	for {
		s.Handler.Handle(zio.NewReadWriter(c))
	}
}

func (s *Server) handleConnZmux(conn net.Conn) {
	mux := zmux.New(conn)
	for {
		ch, err := mux.Accept()
		if err != nil {
			log.Printf("failed to accept channel: %s\n", err)
			continue
		}
		go s.handleChannelZmux(ch)
	}
}

func (s *Server) handleChannelZmux(ch zmux.Channel) {
	for {
		s.Handler.Handle(zio.NewReadWriter(ch))
	}
}
