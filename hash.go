package yamf

import (
	"bytes"
	"errors"
	"hash"
	"sync"
)

var (
	ErrHashNotFound = errors.New("hash type not found")
)

type hashRegistry struct {
	mu     sync.RWMutex
	hashes map[uint64]func() (Hash, error)
}

func (h *hashRegistry) registerHash(id uint64, f func() (Hash, error)) {
	h.mu.Lock()
	if h.hashes == nil {
		h.hashes = make(map[uint64]func() (Hash, error))
	}
	h.hashes[id] = f
	h.mu.Unlock()
}

func (h *hashRegistry) createHash(id uint64) (Hash, error) {
	h.mu.RLock()
	f, ok := h.hashes[id]
	h.mu.RUnlock()
	if !ok {
		return nil, ErrHashNotFound
	}
	return f()
}

func (h *hashRegistry) typeValueToHash(val TypeValue) (Hash, error) {
	return h.createHash(val.Type)
}

var registry hashRegistry

// RegisterHash associates factory function with an hash type identifier.
// If a given type id exists, the function will be overwritten.
//
// RegisterHash can be called concurrently.
func RegisterHash(id uint64, f func() (Hash, error)) {
	registry.registerHash(id, f)
}

type Hash interface {
	hash.Hash
	HashType() uint64
}

func HashToTypeValue(hash Hash) *TypeValue {
	return &TypeValue{
		Type:  hash.HashType(),
		Value: hash.Sum(nil),
	}
}

func TypeValueToHash(val TypeValue) (Hash, error) {
	return registry.typeValueToHash(val)
}

// CreateHash creates hash object associated with given type id
// and an error, if any.
//
// CreateHash can be called concurrently.
func CreateHash(id uint64) (Hash, error) {
	return registry.createHash(id)
}

// HashEqual is a handy function to compare a Hash object's sum
// to a sum stored in TypeValue.
func HashEqual(a Hash, b TypeValue) bool {
	if a.HashType() != b.Type {
		return false
	}
	return bytes.Equal(a.Sum(nil), b.Value)
}
