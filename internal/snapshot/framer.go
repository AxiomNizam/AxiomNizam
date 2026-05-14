// Package snapshot provides a length-prefixed, CRC-checked frame
// format for streaming snapshots between peers.
//
// The format is a simple loop of records:
//
//	+--------+--------+------------------+
//	| u32 len| u32 crc| payload (len B)  |
//	+--------+--------+------------------+
//
// length and crc are big-endian.  crc is CRC32-IEEE over payload.
// A length of 0 is the end-of-stream sentinel; the reader must
// verify its crc is also 0.
//
// The writer buffers nothing — each Write emits one frame — so that
// progress is observable over slow network links.  This mirrors the
// Nomad/Consul approach where snapshots can be many gigabytes and
// the restore side wants to fsync as data arrives.
package snapshot

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
)

// ErrCRC is returned by Reader when a frame fails CRC validation.
var ErrCRC = errors.New("snapshot: CRC mismatch")

// ErrFraming is returned when a frame's length field is impossible
// (exceeds the configured max) or the EOF sentinel has a non-zero
// CRC.
var ErrFraming = errors.New("snapshot: framing error")

// MaxFrame caps one payload at 16 MiB.  Large enough for any
// reasonable log chunk; small enough to bound the reader's allocation.
const MaxFrame = 16 * 1024 * 1024

// Writer emits frames to an underlying io.Writer.
type Writer struct {
	w   io.Writer
	buf [8]byte
}

// NewWriter wraps w.
func NewWriter(w io.Writer) *Writer { return &Writer{w: w} }

// WriteFrame emits one payload frame.  Zero-length payloads are
// forbidden — call Close to emit the EOF sentinel instead.
func (w *Writer) WriteFrame(p []byte) error {
	if len(p) == 0 {
		return errors.New("snapshot: empty frame — use Close")
	}
	if len(p) > MaxFrame {
		return ErrFraming
	}
	binary.BigEndian.PutUint32(w.buf[0:4], uint32(len(p)))
	binary.BigEndian.PutUint32(w.buf[4:8], crc32.ChecksumIEEE(p))
	if _, err := w.w.Write(w.buf[:]); err != nil {
		return err
	}
	_, err := w.w.Write(p)
	return err
}

// Close emits the end-of-stream sentinel (len=0, crc=0).  Callers
// are responsible for closing the underlying Writer.
func (w *Writer) Close() error {
	var eos [8]byte
	_, err := w.w.Write(eos[:])
	return err
}

// Reader parses frames from an underlying io.Reader.
type Reader struct {
	r   io.Reader
	buf [8]byte
	// done is set after the EOF sentinel is consumed — further Next
	// calls will return io.EOF.
	done bool
}

// NewReader wraps r.
func NewReader(r io.Reader) *Reader { return &Reader{r: r} }

// Next returns the next frame's payload.  Returns (nil, io.EOF)
// when the EOF sentinel has been consumed.  The returned slice is
// freshly allocated; callers may retain it.
func (r *Reader) Next() ([]byte, error) {
	if r.done {
		return nil, io.EOF
	}
	if _, err := io.ReadFull(r.r, r.buf[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(r.buf[0:4])
	want := binary.BigEndian.Uint32(r.buf[4:8])
	if length == 0 {
		if want != 0 {
			return nil, ErrFraming
		}
		r.done = true
		return nil, io.EOF
	}
	if length > MaxFrame {
		return nil, ErrFraming
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(r.r, payload); err != nil {
		return nil, err
	}
	if crc32.ChecksumIEEE(payload) != want {
		return nil, ErrCRC
	}
	return payload, nil
}
