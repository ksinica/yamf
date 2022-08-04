package varu64_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/ksinica/yamf/varu64"
)

func TestUint64String(t *testing.T) {
	cases := map[uint64]string{
		0:                 "0",
		72057594037927936: "72057594037927936",
	}
	for input, expected := range cases {
		val := varu64.U64(input)
		if val.String() != expected {
			t.Fatalf("Expected %v, got %v\n", expected, val)
		}
	}
}

func TestUint64BinaryMarshalUnmarshal(t *testing.T) {
	cases := map[uint64][]byte{
		0:                    {0},
		1:                    {1},
		247:                  {247},
		248:                  {248, 248},
		255:                  {248, 255},
		256:                  {249, 1, 0},
		65535:                {249, 255, 255},
		65536:                {250, 1, 0, 0},
		72057594037927935:    {254, 255, 255, 255, 255, 255, 255, 255},
		72057594037927936:    {255, 1, 0, 0, 0, 0, 0, 0, 0},
		18446744073709551615: {255, 255, 255, 255, 255, 255, 255, 255, 255},
	}
	for input, expected := range cases {
		val := varu64.U64(input)
		got, err := val.MarshalBinary()
		if err != nil {
			t.Fatalf("MarshalBinary failed for %d: %s\n", input, err)
		}
		if !bytes.Equal(got, expected) {
			t.Fatalf(
				"Expected encoding for %d: %v, got %v\n",
				input,
				expected,
				got,
			)
		}

		val.SetValue(0)
		if err := val.UnmarshalBinary(got); err != nil {
			t.Fatalf("UnmarshalBinary failed for %d: %s\n", input, err)
		}
		if val.Value() != input {
			t.Fatalf("Expected %d, but decoded %v\n", input, val)
		}
	}
}

func TestUint64UnmarshalTextError(t *testing.T) {
	cases := map[string]error{
		"":                  io.ErrUnexpectedEOF,
		"f9":                io.ErrUnexpectedEOF,
		"f901":              io.ErrUnexpectedEOF,
		"fffffffffffffffff": io.ErrUnexpectedEOF,
		"f82a":              varu64.ErrNonCanonical,
		"f9002a":            varu64.ErrNonCanonical,
	}
	for input, expected := range cases {
		var val varu64.Uint64
		if err := val.UnmarshalText([]byte(input)); !errors.Is(err, expected) {
			t.Fatalf(
				"Expected error for %s: %s, got %s\n",
				input,
				expected,
				err,
			)
		}
	}
}

func TestUint64JsonMarshalUnmarshal(t *testing.T) {
	cases := map[varu64.Uint64]string{
		varu64.U64(0):                 `"00"`,
		varu64.U64(72057594037927935): `"feffffffffffffff"`,
	}
	for input, expected := range cases {
		b, err := json.Marshal(&input)
		if err != nil {
			t.Fatalf("json.Marshal failed: %s\n", err)
		}
		if string(b) != expected {
			t.Fatalf("Expected %s, got %s\n", expected, string(b))
		}

		var val varu64.Uint64
		if err := json.Unmarshal(b, &val); err != nil {
			t.Fatalf("json.Unmarshal failed: %s\n", err)
		}
		if val.Value() != input.Value() {
			t.Fatalf("Expected %v, got %v\n", input, val)
		}
	}
}

func TestUint64WriteToReadFromCounts(t *testing.T) {
	cases := map[uint64]int64{
		247:               1,
		255:               2,
		65535:             3,
		16777215:          4,
		4294967295:        5,
		1099511627775:     6,
		281474976710655:   7,
		72057594037927935: 8,
		72057594037927936: 9,
	}
	for input, expected := range cases {
		var b bytes.Buffer
		val := varu64.U64(input)
		n, err := val.WriteTo(&b)
		if err != nil {
			t.Fatalf("WriteTo failed: %s\n", err)
		}
		if n != expected {
			t.Errorf("Expected to write %d, written %d\n", expected, n)
		}

		n, err = val.ReadFrom(&b)
		if err != nil {
			t.Fatalf("ReadFrom failed: %s\n", err)
		}
		if n != expected {
			t.Errorf("Expected to read %d, read %d\n", expected, n)
		}
	}
}
