package zmux

import (
	"io"
	"sync/atomic"
)

type Channel interface {
	io.ReadWriteCloser
	SetFrameSize(frameSize uint32)
}

func makeChannel(id uint16, rb *LimitBuffer, w io.Writer, frameSize uint32) *channel {
	ch := &channel{id: id, rb: rb, w: w}
	ch.SetFrameSize(frameSize)
	return ch
}

type channel struct {
	id uint16
	rb *LimitBuffer
	w  io.Writer

	frameSize atomic.Uint32
}

func (c *channel) SetFrameSize(frameSize uint32) {
	c.frameSize.Store(frameSize)
}

func (c *channel) Read(b []byte) (n int, err error) {
	return c.rb.Read(b)
}

// можно в канале делить на фреймы. Это несущественно.
// Просто на 1 фрейм больше будет слаться изредка, пофиг.
func (c *channel) Write(b []byte) (n int, err error) {
	fs := min(c.frameSize.Load(), uint32(len(b)))
	buf := make([]byte, 0, len(Header{})+int(fs))

	start := uint32(0)
	end := uint32(0)
	for int(end) < len(b) {
		start = end
		end = min(end+fs, uint32(len(b)))
		h := NewHeader(PAYLOAD, c.id, end-start)
		buf = append(buf, h[:]...)
		buf = append(buf, b[start:end]...)
		n, err := c.w.Write(buf)
		if err != nil {
			return int(start) + n, err
		}
		buf = buf[:0]
	}

	return len(b), nil
}

func (c *channel) Close() error {
	// TODO
	return nil
}
