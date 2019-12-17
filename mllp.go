package hl7

// from https://github.com/deoxxa/mllp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

const (
	MSG_PREFIX  = 0x0b
	MSG_SUFFIX  = 0x1c
	MSG_NEWLINE = 0x0d
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

func (r Reader) ReadMessage(strict bool) ([]byte, error) {
	c, err := r.b.ReadByte()
	if err != nil {
		return nil, err
	}

	if c != byte(MSG_PREFIX) {
		if strict {
			return nil, ErrMLLPInvalidHeader(fmt.Errorf("invalid header found; expected %02x; got %02x", MSG_PREFIX, c))
		}
		if err := r.b.UnreadByte(); err != nil {
			return nil, err
		}
	}

	d, err := r.b.ReadBytes(byte(MSG_SUFFIX))
	if d == nil || len(d) < 2 {
		return nil, ErrMLLPInvalidBoundary(fmt.Errorf("content including boundary should be at least two bytes long; was %d", len(d)))
	}
	if strict {
		if err != nil {
			return nil, ErrMLLPMissingTrailer(err)
		}

		if d[len(d)-2] != MSG_NEWLINE {
			return nil, ErrMLLPInvalidContent(fmt.Errorf("content should end with %02x; instead was %02x", MSG_NEWLINE, d[len(d)-2]))
		}

		t, err := r.b.ReadByte()
		if err != nil {
			return nil, err
		}
		if t != byte(MSG_NEWLINE) {
			return nil, ErrMLLPMissingTrailer(fmt.Errorf("invalid trailer found; expected %02x; got %02x", MSG_NEWLINE, t))
		}
		d = d[0 : len(d)-2]
	} else {
		c := bytes.IndexByte(d, byte(MSG_SUFFIX))
		if c != -1 {
			d = d[0 : c-1]
		}
	}

	return d, nil
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w}
}

func (w Writer) WriteMessage(b []byte, withBoundary bool) error {
	if withBoundary {
		if _, err := w.w.Write([]byte{0x0b}); err != nil {
			return err
		}
	}

	for len(b) > 0 {
		n, err := w.w.Write(b)
		if err != nil {
			return err
		}
		b = b[n:]
	}
	if withBoundary {
		if _, err := w.w.Write([]byte{MSG_NEWLINE, MSG_SUFFIX, MSG_NEWLINE}); err != nil {
			return err
		}
	}

	return nil
}

func NewReadWriter(rw io.ReadWriter) *ReadWriter {
	return &ReadWriter{NewReader(rw), NewWriter(rw)}
}
