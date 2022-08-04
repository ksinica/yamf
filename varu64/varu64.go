package varu64

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
)

var (
	// ErrNonCanonical is a non-critical error signalizing that a non-canonical
	// VarU64 encoding was encountered. Operations returning ErrNonCanonical
	// can be considered successful.
	ErrNonCanonical = errors.New("varu64: non-canonical encoding")
)

// Uint64 is a handy data transfer object that wraps uint64 and allows encoding
// and decoding as defined by the VarU64 specification.
type Uint64 struct {
	uint64

	io.ReaderFrom
	io.WriterTo
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

// Returns the string representation of the receiver's underlying uint64 value
// in base 10.
func (v *Uint64) String() string {
	return strconv.FormatUint(v.uint64, 10)
}

// Value returns the receiver's underlying uint64 value.
func (v *Uint64) Value() uint64 {
	return v.uint64
}

// SetValue sets the receiver's underlying uint64 value to value.
func (v *Uint64) SetValue(value uint64) {
	v.uint64 = value
}

// ReadFrom decodes VarU64 from r until EOF or error, and sets the receiver's
// underlying uint64 value to the decoded one. It returns the number
// of bytes read and an error, if any, except EOF.
func (v *Uint64) ReadFrom(r io.Reader) (int64, error) {
	val, n, err := ReadUint64(r)
	if err == nil || errors.Is(err, ErrNonCanonical) {
		v.uint64 = val
	}
	return int64(n), err
}

// WriteTo encodes receiver to w. It returns the number of bytes written
// and an error, if any.
func (v *Uint64) WriteTo(w io.Writer) (int64, error) {
	n, err := WriteUint64(w, v.uint64)
	return int64(n), err
}

// MarshalBinary encodes receiver into a byte slice and returns an error,
// if any.
func (v *Uint64) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	if _, err := v.WriteTo(&b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// UnmarshalBinary decodes VarU64-encoded data into the receiver.
func (v *Uint64) UnmarshalBinary(data []byte) error {
	_, err := v.ReadFrom(bytes.NewBuffer(data))
	return err
}

// MarshalText encodes receiver into a textual representation of
// VarU64 encoding.
func (v *Uint64) MarshalText() ([]byte, error) {
	var b bytes.Buffer
	_, err := WriteUint64(hex.NewEncoder(&b), v.uint64)
	return b.Bytes(), err
}

// UnmarshalText decodes VarU64-encoded textual representation into
// the receiver.
func (v *Uint64) UnmarshalText(text []byte) error {
	val, _, err := ReadUint64(hex.NewDecoder(bytes.NewBuffer(text)))
	if err != nil {
		return err
	}
	v.uint64 = val
	return nil
}

// U64 returns value wrapped as Uint64.
func U64(value uint64) Uint64 {
	return Uint64{uint64: value}
}

func encodingParams(v uint64) (byte, int) {
	switch {
	case v < 248:
		return byte(v), 0
	case v < 256:
		return 248, 1
	case v < 65536:
		return 249, 2
	case v < 16777216:
		return 250, 3
	case v < 4294967296:
		return 251, 4
	case v < 1099511627776:
		return 252, 5
	case v < 281474976710656:
		return 253, 6
	case v < 72057594037927936:
		return 254, 7
	default:
		return 255, 8
	}
}

// WriteUint64 encodes value to w. It returns the number of bytes written
// and an error, if any.
func WriteUint64(w io.Writer, value uint64) (int, error) {
	header, length := encodingParams(value)

	var buf [9]byte
	if length > 0 {
		binary.BigEndian.PutUint64(buf[1:], value)
	}

	i := (len(buf) - length) - 1
	buf[i] = header

	return w.Write(buf[i:])
}

func eofToUnexpectedEOF(err error) error {
	if errors.Is(err, io.EOF) {
		return io.ErrUnexpectedEOF
	}
	return err
}

// ReadFrom decodes VarU64 from r until EOF or error, and returns decoded value,
// number of bytes read and an error, if any, except EOF.
func ReadUint64(r io.Reader) (uint64, int, error) {
	var buf [9]byte
	if n, err := r.Read(buf[:1]); err != nil {
		return 0, n, eofToUnexpectedEOF(err)
	}
	if buf[0]|7 != 255 {
		return uint64(buf[0]), 1, nil
	}

	length := int((buf[0] & 7) + 1)

	if n, err := io.ReadFull(r, buf[len(buf)-length:]); err != nil {
		return 0, n, eofToUnexpectedEOF(err)
	}

	v := binary.BigEndian.Uint64(buf[1:])
	if _, l := encodingParams(v); length != l {
		return v, length + 1, ErrNonCanonical
	}
	return v, length + 1, nil
}
