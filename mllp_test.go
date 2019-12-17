package hl7

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func wrapWithMarkers(b []byte) []byte {
	return append(append([]byte{0x0b}, b...), 0x0d, 0x1c, 0x0d)
}

func TestReadMessage(t *testing.T) {
	a := assert.New(t)

	r := NewReader(bytes.NewReader(wrapWithMarkers([]byte("hello"))))

	m, err := r.ReadMessage(true)
	a.NoError(err)
	if a.NotNil(m) {
		a.Equal([]byte("hello"), m)
	}
}

func TestReadInvalidMessageHeader(t *testing.T) {
	a := assert.New(t)

	r := NewReader(bytes.NewReader([]byte{0x01}))

	m, err := r.ReadMessage(true)
	a.Nil(m)
	a.Error(err)
	_, ok := err.(ErrMLLPInvalidHeader)
	a.True(ok)
	a.Contains(err.Error(), "invalid header found; expected 0x0b")
}

func TestReadInvalidMessageBoundary(t *testing.T) {
	a := assert.New(t)

	r := NewReader(bytes.NewReader([]byte{0x0b, 0x0c, 0x1c, 0x00}))

	m, err := r.ReadMessage(true)
	a.Nil(m)
	a.Error(err)
	_, ok := err.(ErrMLLPInvalidContent)
	a.True(ok)
	a.Contains(err.Error(), "content should end with 0x0d")
}

func TestReadShortMessage(t *testing.T) {
	a := assert.New(t)

	r := NewReader(bytes.NewReader([]byte{0x0b, 0x1c, 0x0d}))

	m, err := r.ReadMessage(true)
	a.Nil(m)
	a.Error(err)
	_, ok := err.(ErrMLLPInvalidBoundary)
	a.True(ok)
	a.Contains(err.Error(), "content including boundary should be at least two bytes long")
}

func TestReadInvalidMessageTrailer(t *testing.T) {
	a := assert.New(t)

	r := NewReader(bytes.NewReader([]byte{0x0b, 0x0d, 0x1c, 0x00}))

	m, err := r.ReadMessage(true)
	a.Nil(m)
	a.Error(err)
	_, ok := err.(ErrMLLPMissingTrailer)
	a.True(ok)
	a.Contains(err.Error(), "invalid trailer found; expected 0x0d")
}

func TestWriter(t *testing.T) {
	a := assert.New(t)

	b := bytes.NewBuffer(nil)

	w := NewWriter(b)

	a.NoError(w.WriteMessage([]byte("hello"), true))
	a.Equal(wrapWithMarkers([]byte("hello")), b.Bytes())
}

func TestReadWriter(t *testing.T) {
	a := assert.New(t)

	r := bytes.NewReader(wrapWithMarkers([]byte("input")))
	w := bytes.NewBuffer(nil)

	brw := bufio.NewReadWriter(bufio.NewReader(r), bufio.NewWriter(w))

	rw := NewReadWriter(brw)

	m, err := rw.ReadMessage(true)
	a.NoError(err)
	if a.NotNil(m) {
		a.Equal([]byte("input"), m)
	}

	a.NoError(rw.WriteMessage([]byte("output"), true))
	a.NoError(brw.Flush())
	a.Equal(wrapWithMarkers([]byte("output")), w.Bytes())
}
