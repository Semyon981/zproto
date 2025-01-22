package zproto

import (
	"fmt"
	"net"
	"sync"

	"github.com/Semyon981/zproto/zio"
	"github.com/hashicorp/yamux"
)

// type Client struct {
// 	mux       zmux.Mux
// 	freeChans *sync.Pool
// }

// func NewClient(conn net.Conn) (Client, error) {
// 	return Client{mux: zmux.New(conn), freeChans: new(sync.Pool)}, nil
// }

// func (c *Client) Session() (Session, error) {
// 	ch, ok := c.freeChans.Get().(zmux.Channel)
// 	if !ok {
// 		var err error
// 		ch, err = c.mux.Open()
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to create session: %w", err)
// 		}
// 	}

// 	return newSession(ch, func() {
// 		c.freeChans.Put(ch)
// 	}), nil
// }

// type Session interface {
// 	zio.ReadWriter
// 	io.Closer
// 	// возвращает базовый канал
// 	Channel() zmux.Channel
// }

// func newSession(ch zmux.Channel, close func()) Session {
// 	s := &session{ch: ch, ReadWriter: zio.NewReadWriter(ch), close: close}
// 	return s
// }

// type session struct {
// 	zio.ReadWriter
// 	ch        zmux.Channel
// 	mu        *sync.RWMutex
// 	freeChans []zmux.Channel
// 	closeOnce sync.Once
// 	close     func()
// }

// func (s *session) Channel() zmux.Channel {
// 	return s.ch
// }

// func (s *session) Close() error {
// 	s.closeOnce.Do(s.close)
// 	return nil
// }

type Client interface {
	Session() (zio.ReadWriteCloser, error)
}

type client struct {
	mux       *yamux.Session
	freeChans sync.Pool
}

func NewClient(conn net.Conn) (Client, error) {
	mux, err := yamux.Client(conn, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create mux: %w", err)
	}
	return &client{mux: mux, freeChans: sync.Pool{}}, nil
}

func (c *client) Session() (zio.ReadWriteCloser, error) {
	ch, ok := c.freeChans.Get().(net.Conn)
	if !ok {
		var err error
		ch, err = c.mux.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
	}

	return newSession(ch, func() {
		c.freeChans.Put(ch)
	}), nil
}

func newSession(ch net.Conn, close func()) *session {
	s := &session{ReadWriter: zio.NewReadWriter(ch), close: close}
	return s
}

type session struct {
	zio.ReadWriter
	closeOnce sync.Once
	close     func()
}

func (s *session) Close() error {
	s.closeOnce.Do(s.close)
	return nil
}
