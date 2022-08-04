//go:build go1.18
// +build go1.18

package yamf_test

import (
	"testing"

	"github.com/ksinica/yamf"
)

func FuzzTypeValueBinaryMarshalUnmarshal(f *testing.F) {
	for _, data := range [][]byte{
		{},
		{0x00, 0x00},
		{0x01, 0x04, 0x74, 0x65, 0x73, 0x74},
		{0x02, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		{0x03},
		{0x04, 0x03, 0x01, 0x02},
	} {
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		t.Parallel()
		var val yamf.TypeValue
		if err := val.UnmarshalBinary(input); err == nil {
			b, err := val.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary failed for %v: %s\n", val, err)
			}

			if err := val.UnmarshalBinary(b); err != nil {
				t.Fatalf("UnmarshalBinary failed for: %v: %s\n", b, err)
			}
		}
	})
}
