package blake2b

import (
	"hash"

	"github.com/ksinica/yamf"
	"golang.org/x/crypto/blake2b"
)

const (
	Type         = 0
	DigestLength = 64
)

func init() {
	yamf.RegisterHash(Type, New)
}

type blake2bHash struct {
	hash.Hash
}

func (*blake2bHash) HashType() uint64 {
	return Type
}

func New() (yamf.Hash, error) {
	hash, err := blake2b.New(DigestLength>>3, nil)
	if err != nil {
		return nil, err
	}
	return &blake2bHash{
		Hash: hash,
	}, nil
}
