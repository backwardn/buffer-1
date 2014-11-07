package buffer

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
)

type Memory struct {
	N int64
	*bytes.Buffer
}

func New(n int64) *Memory {
	return &Memory{
		N:      n,
		Buffer: bytes.NewBuffer(nil),
	}
}

func (buf *Memory) Cap() int64 {
	return buf.N
}

func (buf *Memory) Len() int64 {
	return int64(buf.Buffer.Len())
}

func (buf *Memory) Write(p []byte) (n int, err error) {
	return LimitWriter(buf.Buffer, Gap(buf)).Write(p)
}

// Must Add io.ErrShortWrite
func (buf *Memory) WriteAt(p []byte, off int64) (n int, err error) {
	if off > buf.Len() {
		return 0, bytes.ErrTooLarge
	} else if len64(p)+off <= buf.Len() {
		d := buf.Bytes()[off:]
		return copy(d, p), nil
	} else {
		d := buf.Bytes()[off:]
		n = copy(d, p)
		m, err := buf.Write(p[n:])
		return n + m, err
	}
}

func (buf *Memory) ReadAt(p []byte, off int64) (n int, err error) {
	return bytes.NewReader(buf.Bytes()).ReadAt(p, off)
}

func (buf *Memory) Read(p []byte) (n int, err error) {
	return io.LimitReader(buf.Buffer, buf.Len()).Read(p)
}

func (buf *Memory) FFwd(n int64) int64 {
	data := buf.Bytes()
	if n > len64(data) {
		n = len64(data)
	}
	b := bytes.NewBuffer(data[n:])
	buf.Buffer = b
	return n
}

func init() {
	gob.Register(&Memory{})
}

func (buf *Memory) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	fmt.Fprint(&b, buf.N)
	b.Write(buf.Bytes())
	return b.Bytes(), nil
}

func (buf *Memory) UnmarshalBinary(data []byte) error {
	b := bytes.NewBuffer(data)
	_, err := fmt.Fscan(b, &buf.N)
	buf.Buffer = b
	return err
}
