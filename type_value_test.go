package yamf_test

import (
	"bytes"
	"errors"
	"io"
	"math"
	"testing"

	"github.com/ksinica/yamf"
	"github.com/ksinica/yamf/varu64"
)

func TestBinaryMarshalUnmarshal(t *testing.T) {
	cases := []struct {
		input    yamf.TypeValue
		expected []byte
	}{
		{
			input:    yamf.TypeValue{Type: 0},
			expected: []byte{0x00, 0x00},
		},
		{
			input:    yamf.TypeValue{Type: 1, Value: []byte("test")},
			expected: []byte{0x01, 0x04, 0x74, 0x65, 0x73, 0x74},
		},
	}
	for _, x := range cases {
		b, err := x.input.MarshalBinary()
		if err != nil {
			t.Errorf("MarshalBinary failed for %v: %s\n", x.input, err)
		}

		if !bytes.Equal(x.expected, b) {
			t.Errorf("Expected %s, got %s\n", x.expected, b)
		}

		var tv yamf.TypeValue
		if err := tv.UnmarshalBinary(b); err != nil {
			t.Errorf("UnmarshalBinary failed: %s\n", err)
		}

		if !tv.Equal(x.input) {
			t.Errorf("Expected %v, got %v\n", x.input, tv)
		}
	}
}

func makeTooBigTypeValue(t *testing.T) string {
	val := yamf.TypeValue{
		Type:  math.MaxUint64,
		Value: make([]byte, yamf.MaxValueSize+1),
	}

	b, err := val.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestTextMarshalUnmarshalFail(t *testing.T) {
	cases := map[string]error{
		"":                     io.ErrUnexpectedEOF,
		"00":                   io.ErrUnexpectedEOF,
		"0001":                 io.ErrUnexpectedEOF,
		"000201":               io.ErrUnexpectedEOF,
		"f82a0101":             varu64.ErrNonCanonical,
		makeTooBigTypeValue(t): yamf.ErrValueTooBig,
	}
	for input, expected := range cases {
		var val yamf.TypeValue
		if err := val.UnmarshalText([]byte(input)); !errors.Is(err, expected) {
			t.Fatalf("Expected for \"%s\": %s, got %s\n", input, expected, err)
		}
	}
}
