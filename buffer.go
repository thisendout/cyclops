package main

import (
	"bytes"
	"io"
)

// implements io.Writer to receive streaming logs
// currently a passthrough to bytes.Buffer but can
// also attach a channel to stream to other readers
type Buffer struct {
	buf    bytes.Buffer
	writer io.Writer
}

func NewBuffer(w io.Writer) *Buffer {
	return &Buffer{writer: w}
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	b.writer.Write(p)
	n, err = b.buf.Write(p)
	return
}

func (b *Buffer) WriteString(s string) (n int, err error) {
	p := []byte(s)
	b.writer.Write(p)
	n, err = b.buf.WriteString(s)
	return
}

func (b *Buffer) Bytes() []byte {
	return b.buf.Bytes()
}
