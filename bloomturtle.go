package bloomturtle

import (
	"github.com/clarencemoreno/bloomturtle/internal/bloomfilter"
)

// NewBloomFilter creates a new Bloom filter
func NewBloomFilter(size uint, hashFuncs []func([]byte) uint) *bloomfilter.BloomFilter {
	return bloomfilter.New(size, hashFuncs)
}
