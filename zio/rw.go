package zio

import (
	"io"
)

type ReadWriter interface {
	io.ReadWriter
	ReadUint32() (uint32, error)
	WriteUint32(v uint32) error
	ReadBytes() ([]byte, error)
	WriteBytes(b []byte) error
	ReadString() (string, error)
	WriteString(s string) error
	ReadJson(v any) error
	WriteJson(v any) error
}

func NewReadWriter(rw io.ReadWriter) ReadWriter {
	return &readWriter{ReadWriter: rw}
}

type readWriter struct {
	io.ReadWriter
}

func (rw *readWriter) ReadUint32() (uint32, error) {
	return ReadUint32(rw)
}

func (rw *readWriter) WriteUint32(v uint32) error {
	return WriteUint32(rw, v)
}

func (rw *readWriter) ReadBytes() ([]byte, error) {
	return ReadBytes(rw)
}

func (rw *readWriter) WriteBytes(b []byte) error {
	return WriteBytes(rw, b)
}

func (rw *readWriter) ReadString() (string, error) {
	return ReadString(rw)
}

func (rw *readWriter) WriteString(s string) error {
	return WriteString(rw, s)
}

func (rw *readWriter) ReadJson(v any) error {
	return ReadJson(rw, v)
}

func (rw *readWriter) WriteJson(v any) error {
	return WriteJson(rw, v)
}

type ReadWriteCloser interface {
	ReadWriter
	io.Closer
}

func NewReadWriteCloser(rw io.ReadWriter, c io.Closer) ReadWriteCloser {
	rrw, ok := rw.(ReadWriter)
	if !ok {
		rrw = NewReadWriter(rw)
	}
	return &readWriteCloser{ReadWriter: rrw, Closer: c}
}

type readWriteCloser struct {
	ReadWriter
	io.Closer
}
