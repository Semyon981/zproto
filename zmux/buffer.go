package zmux

import (
	"io"
	"sync"
	"sync/atomic"
)

func NewLimitBuffer(size int) *LimitBuffer {
	return &LimitBuffer{
		buf:  make([]byte, size),
		cond: sync.NewCond(&sync.Mutex{}),
		wch:  make(chan struct{}, 1),
		rch:  make(chan struct{}, 1),
	}
}

type LimitBuffer struct {
	buf   []byte
	cond  *sync.Cond
	start int
	sz    int

	wch chan struct{}
	rch chan struct{}

	writers atomic.Uint64
}

func (b *LimitBuffer) Read(p []byte) (int, error) {
	b.rch <- struct{}{}
	b.cond.L.Lock()

	for (b.sz < len(b.buf) && b.writers.Load() > 0) || b.sz == 0 {
		b.cond.Wait()
	}

	n := min(len(p), b.sz)
	end := b.start + n
	if end <= len(b.buf) {
		copy(p, b.buf[b.start:end])
	} else {
		copy(p, b.buf[b.start:])
		copy(p[len(b.buf)-b.start:], b.buf[:end-len(b.buf)])
	}

	b.start = (b.start + n) % len(b.buf)
	b.sz -= n

	b.cond.Signal()

	b.cond.L.Unlock()
	<-b.rch
	return n, nil
}

func (b *LimitBuffer) Write(p []byte) (int, error) {
	b.writers.Add(1)
	b.wch <- struct{}{}
	b.cond.L.Lock()

	n := 0
	for n != len(p) {
		for len(b.buf) == b.sz {
			b.cond.Signal()
			b.cond.Wait()
		}

		toWrite := min(len(b.buf)-b.sz, len(p)-n)
		writePos := (b.start + b.sz) % len(b.buf)

		if writePos+toWrite <= len(b.buf) {
			copy(b.buf[writePos:], p[n:n+toWrite])
		} else {
			fpidx := n + len(b.buf) - writePos
			copy(b.buf[writePos:], p[n:fpidx])
			copy(b.buf, p[fpidx:n+toWrite])
		}

		b.sz += toWrite
		n += toWrite
	}

	if b.writers.Add(^uint64(0)) == 0 && b.sz > 0 {
		b.cond.Signal()
	}

	b.cond.L.Unlock()
	<-b.wch
	return n, nil
}

// var cntTrue int
// var cntAll int

func (b *LimitBuffer) WriteTo(w io.Writer) (int64, error) {
	b.rch <- struct{}{}
	b.cond.L.Lock()

	for (b.sz < len(b.buf) && b.writers.Load() > 0) || b.sz == 0 {
		b.cond.Wait()

		// if (b.sz < len(b.buf) && b.writers.Load() > 0) || b.sz == 0 {
		// 	cntTrue++
		// }
		// cntAll++

		// if cntAll%1000 == 0 {
		// 	fmt.Println(cntAll, cntTrue, cntAll-cntTrue)
		// }
	}

	end := b.start + b.sz
	var writed int64
	var err error
	var n int
	if end <= len(b.buf) {
		n, err = w.Write(b.buf[b.start:end])
		writed, b.sz = writed+int64(n), b.sz-n
	} else {
		n, err := w.Write(b.buf[b.start:])
		writed, b.sz = writed+int64(n), b.sz-n
		if err == nil && b.start != 0 {
			n, err = w.Write(b.buf[:end%len(b.buf)])
			writed, b.sz = writed+int64(n), b.sz-n
		}
	}

	b.cond.Signal()
	b.cond.L.Unlock()
	<-b.rch
	return writed, err
}

func (b *LimitBuffer) ReadFrom(r io.Reader) (int64, error) {
	b.writers.Add(1)
	b.wch <- struct{}{}
	b.cond.L.Lock()

	var readed int64
	var err error
	var n int
	for err == nil {
		for len(b.buf) == b.sz {
			b.cond.Signal()
			b.cond.Wait()
		}
		freeStart := (b.start + b.sz) % len(b.buf)
		freeEnd := freeStart + len(b.buf) - b.sz

		n, err = io.ReadFull(r, b.buf[freeStart:min(len(b.buf), freeEnd)])
		readed, b.sz = readed+int64(n), b.sz+n
		if err == nil && freeEnd > len(b.buf) {
			n, err = io.ReadFull(r, b.buf[:freeEnd%len(b.buf)])
			readed, b.sz = readed+int64(n), b.sz+n
		}
	}

	if err == io.EOF || err == io.ErrUnexpectedEOF {
		err = nil
	}

	if b.writers.Add(^uint64(0)) == 0 && b.sz > 0 {
		b.cond.Signal()
	}
	b.cond.L.Unlock()
	<-b.wch
	return readed, err
}
