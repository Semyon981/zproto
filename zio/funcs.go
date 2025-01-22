package zio

import (
	"encoding/binary"
	"encoding/json"
	"io"
)

func ReadUint32(r io.Reader) (uint32, error) {
	b := make([]byte, 4)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b), nil
}

func WriteUint32(w io.Writer, v uint32) error {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	_, err := w.Write(b)
	return err
}

func ReadBytes(r io.Reader) ([]byte, error) {
	sz, err := ReadUint32(r)
	if err != nil {
		return nil, err
	}
	b := make([]byte, sz)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	return b, err
}

func WriteBytes(w io.Writer, b []byte) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(len(b)))
	_, err := w.Write(append(buf, b...))
	return err
}

func ReadString(r io.Reader) (string, error) {
	b, err := ReadBytes(r)
	return string(b), err
}

func WriteString(w io.Writer, s string) error {
	return WriteBytes(w, []byte(s))
}

func ReadJson(r io.Reader, v any) error {
	data, err := ReadBytes(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func WriteJson(w io.Writer, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return WriteBytes(w, b)
}
