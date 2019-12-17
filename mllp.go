package hl7

// from https://github.com/deoxxa/mllp

import (
	"bufio"
	"fmt"
	"io"
)

type (
	ErrMLLPInvalidHeader   error
	ErrMLLPMissingTrailer  error
	ErrMLLPInvalidBoundary error
	ErrMLLPInvalidContent  error

	Reader struct {
		b *bufio.Reader
	}

	Writer struct {
		w io.Writer
	}

	ReadWriter struct {
		*Reader
		*Writer
	}
)

func NewReader(r io.Reader) *Reader {
	return &Reader{bufio.NewReader(r)}
}

func (r Reader) ReadMessage() ([]byte, error) {
	c, err := r.b.ReadByte()
	if err != nil {
		return nil, err
	}

	if c != byte(0x0b) {
		return nil, ErrMLLPInvalidHeader(fmt.Errorf("invalid header found; expected 0x0b; got %02x", c))
	}

	d, err := r.b.ReadBytes(byte(0x1c))
	if err != nil {
		return nil, ErrMLLPMissingTrailer(err)
	}
	if len(d) < 2 {
		return nil, ErrMLLPInvalidBoundary(fmt.Errorf("content including boundary should be at least two bytes long; was %d", len(d)))
	}

	if d[len(d)-2] != 0x0d {
		return nil, ErrMLLPInvalidContent(fmt.Errorf("content should end with 0x0d; instead was %02x", d[len(d)-2]))
	}

	t, err := r.b.ReadByte()
	if err != nil {
		return nil, err
	}
	if t != byte(0x0d) {
		return nil, ErrMLLPMissingTrailer(fmt.Errorf("invalid trailer found; expected 0x0d; got %02x", t))
	}

	return d[0 : len(d)-2], nil
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w}
}

func (w Writer) WriteMessage(b []byte) error {
	if _, err := w.w.Write([]byte{0x0b}); err != nil {
		return err
	}

	for len(b) > 0 {
		n, err := w.w.Write(b)
		if err != nil {
			return err
		}
		b = b[n:]
	}
	if _, err := w.w.Write([]byte{0x0d, 0x1c, 0x0d}); err != nil {
		return err
	}

	return nil
}

func NewReadWriter(rw io.ReadWriter) *ReadWriter {
	return &ReadWriter{NewReader(rw), NewWriter(rw)}
}
