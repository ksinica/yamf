package yamf

import (
	"bytes"
	"encoding"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/ksinica/yamf/varu64"
)

var (
	// MaxValueSize sets the maximal value byte size for read operations.
	// Less than zero means no limit.
	MaxValueSize = 1024
)

var (
	// ErrValueTooBig is returned when value byte size exceeds limit
	// set by MaxValueSize.
	ErrValueTooBig = errors.New("yamf: value is too big")
)

// TypeValue encapsulates type and value pair and allows to encode and decode
// them as per Simple Type-Length-Value specification:
//     https://github.com/AljoschaMeyer/stlv
type TypeValue struct {
	Type  uint64
	Value []byte

	io.ReaderFrom
	io.WriterTo
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

func (t TypeValue) String() string {
	return fmt.Sprintf("{Type: %d, Value: %v}", t.Type, t.Value)
}

func (t TypeValue) Equal(other TypeValue) bool {
	return t.Type == other.Type && bytes.Equal(t.Value, other.Value)
}

func (t *TypeValue) writeValue(w io.Writer) (n int64, err error) {
	if len(t.Value) > 0 {
		count, err := w.Write(t.Value)
		n += int64(count)
		if err != nil {
			return n, err
		}
	}
	return
}

// WriteTo encodes receiver to w. It returns the number of bytes written
// and an error, if any.
func (t *TypeValue) WriteTo(w io.Writer) (n int64, err error) {
	var val varu64.Uint64

	val.SetValue(t.Type)
	count, err := val.WriteTo(w)
	n += count
	if err != nil {
		return n, err
	}

	val.SetValue(uint64(len(t.Value)))
	count, err = val.WriteTo(w)
	n += count
	if err != nil {
		return n, err
	}

	count, err = t.writeValue(w)
	n += count
	return n, err
}

func readFull(r io.Reader, p []byte) (n int, err error) {
	n, err = io.ReadFull(r, p)
	if n == 0 && errors.Is(err, io.EOF) {
		err = io.ErrUnexpectedEOF
	}
	return
}

func (t *TypeValue) readValue(r io.Reader, length int) (int64, error) {
	if length > 0 {
		if MaxValueSize > 0 && length > MaxValueSize {
			return 0, ErrValueTooBig
		}

		t.Value = make([]byte, length)

		n, err := readFull(r, t.Value)
		return int64(n), err
	}
	return 0, nil
}

// MarshalBinary encodes receiver into a byte slice and returns an error,
// if any.
//
// varu64.ErrNonCanonical may be returned in the case when non-canonical
// encoding was encountered during the process.
func (t *TypeValue) ReadFrom(r io.Reader) (n int64, err error) {
	var val varu64.Uint64
	var errs [3]error

	count, err := val.ReadFrom(r)
	n += count
	if err != nil && !errors.Is(err, varu64.ErrNonCanonical) {
		return n, err
	}
	errs[0] = err
	t.Type = val.Value()

	count, err = val.ReadFrom(r)
	n += count
	if err != nil && !errors.Is(err, varu64.ErrNonCanonical) {
		return n, err
	}
	errs[1] = err

	if val.Value() > math.MaxInt {
		return n, ErrValueTooBig
	}

	count, err = t.readValue(r, int(val.Value()))
	n += count
	errs[2] = err

	for i := 2; i >= 0; i-- {
		if err = errs[i]; err != nil {
			break
		}
	}
	return
}

// MarshalBinary encodes receiver into a byte slice and returns an error,
// if any.
func (s *TypeValue) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	if _, err := s.WriteTo(&b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// UnmarshalBinary decodes data into the receiver and returns an error,
// if any.
func (s *TypeValue) UnmarshalBinary(data []byte) error {
	_, err := s.ReadFrom(bytes.NewBuffer(data))
	return err
}

// MarshalText encodes receiver into a textual representation and returns
// an error, if any.
func (s *TypeValue) MarshalText() ([]byte, error) {
	var b bytes.Buffer
	if _, err := s.WriteTo(hex.NewEncoder(&b)); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// UnmarshalText decodes textual representation into the receiver and returns
// an error, if any.
func (s *TypeValue) UnmarshalText(text []byte) error {
	_, err := s.ReadFrom(hex.NewDecoder(bytes.NewBuffer(text)))
	return err
}
