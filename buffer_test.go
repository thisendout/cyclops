package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferWriteString(t *testing.T) {
	assert := assert.New(t)

	var writer bytes.Buffer
	buf := NewBuffer(&writer)

	n, err := buf.WriteString("banana banana")
	assert.NoError(err)
	assert.Equal(n, 13)

	byt := buf.Bytes()
	writerbyt := writer.Bytes()
	assert.Equal(string(byt), "banana banana")
	assert.Equal(string(writerbyt), "banana banana")
}

func TestBufferWrite(t *testing.T) {
	assert := assert.New(t)

	var writer bytes.Buffer
	buf := NewBuffer(&writer)
	p := []byte("banana banana")

	n, err := buf.Write(p)
	assert.NoError(err)
	assert.Equal(n, 13)

	byt := buf.Bytes()
	writerbyt := writer.Bytes()
	assert.Equal(string(byt), "banana banana")
	assert.Equal(string(writerbyt), "banana banana")
}
