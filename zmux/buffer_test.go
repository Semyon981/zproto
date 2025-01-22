package zmux_test

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
	"time"

	"github.com/Semyon981/zproto/zmux"
)

func TestLimitBuffer(t *testing.T) {
	buf := zmux.NewLimitBuffer(1024)

	inp := make([]byte, 128)
	_, err := rand.Read(inp)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		n, err := buf.Write(inp)
		if err != nil {
			t.Error("Write error", err)
		}

		if n != len(inp) {
			t.Error("Write() returns unexpected n")
		}
	}()

	out := make([]byte, len(inp))
	_, err = io.ReadFull(buf, out)
	if err != nil {
		t.Error("Read error", err)
	}

	if !bytes.Equal(inp, out) {
		t.Errorf("data is not equal!")
	}
}

func TestMultiWriting(t *testing.T) {
	bsz := 256
	b := zmux.NewLimitBuffer(bsz)

	for i := 0; i < bsz*2; i++ {
		go func() {
			_, err := b.Write([]byte{0})
			if err != nil {
				t.Error("Write error", err)
			}
		}()
	}
	time.Sleep(time.Millisecond * 100)

	out := make([]byte, bsz*2)
	for i := 0; i < 2; i++ {
		n, err := b.Read(out)
		if err != nil {
			t.Error("Read error", err)
		}

		if n != bsz {
			t.Error("unexpected n", n)
		}
	}
}

func BenchmarkLimitBuffer(t *testing.B) {
	buf := zmux.NewLimitBuffer(1024)

	inp := make([]byte, 100*1024*1024)
	_, err := rand.Read(inp)
	if err != nil {
		t.Fatal(err)
	}

	go func(buf *zmux.LimitBuffer, inp []byte) {
		for i := 0; i < t.N; i++ {
			n, err := buf.Write(inp)
			if err != nil {
				t.Error("Write error", err)
			}

			if n != len(inp) {
				t.Error("Write() returns unexpected n")
			}
		}
	}(buf, inp)

	time.Sleep(time.Second)

	out := make([]byte, len(inp))

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		t.StartTimer()
		start := time.Now()
		n, err := io.ReadFull(buf, out)
		if err != nil {
			t.Error("Read error", err)
		}
		if n != len(out) {
			t.Error("ReadFull returns unexpected n")
		}
		t.Log(time.Since(start))
		t.StopTimer()

		if !bytes.Equal(inp, out) {
			t.Errorf("data is not equal!")
		}
	}

}

func BenchmarkSimpleCopy(t *testing.B) {
	inp := make([]byte, 100*1024*1024)
	_, err := rand.Read(inp)
	if err != nil {
		t.Fatal(err)
	}
	out := make([]byte, len(inp))

	arr := make([]byte, 1024)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		start := time.Now()
		t.StartTimer()
		for i := 0; i < len(inp)/len(arr); i++ {
			copy(arr, inp[i*len(arr):])
			copy(out[i*len(arr):], arr)
		}
		t.StopTimer()
		t.Log(time.Since(start))
		if !bytes.Equal(inp, out) {
			t.Errorf("data is not equal!")
		}
	}
}
