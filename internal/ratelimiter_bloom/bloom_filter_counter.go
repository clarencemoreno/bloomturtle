package ratelimiter_bloom

import (
	"hash/fnv"
	"sync"
)

// BloomFilterCounter implements a Bloom filter counter.
type BloomFilterCounter struct {
	array    []uint64
	size     uint32
	capacity uint32
	mu       sync.RWMutex
}

// NewBloomFilterCounter creates a new Bloom filter counter with the specified size.
// The default capacity is set to 10 if not specified.
func NewBloomFilterCounter(size uint32, capacity ...uint32) *BloomFilterCounter {
	cap := uint32(10)
	if len(capacity) > 0 {
		cap = capacity[0]
	}
	bfc := &BloomFilterCounter{
		array:    make([]uint64, size),
		size:     size,
		capacity: cap,
	}
	for i := range bfc.array {
		bfc.array[i] = uint64(cap)
	}
	return bfc
}

// Increment increments the counter for the given key by the specified count.
func (bfc *BloomFilterCounter) Increment(key string, count int) {
	index := bfc.hash(key) % uint64(bfc.size)
	bfc.mu.Lock()
	defer bfc.mu.Unlock()
	bfc.array[index] += uint64(count)
}

// Decrement decrements the counter for the given key.
func (bfc *BloomFilterCounter) Decrement(key string) {
	index := bfc.hash(key) % uint64(bfc.size)
	bfc.mu.Lock()
	defer bfc.mu.Unlock()
	if bfc.array[index] > 0 {
		bfc.array[index]--
	}
}

// IsEmpty checks if the counter for the given key is greater than 0.
func (bfc *BloomFilterCounter) IsEmpty(key string) bool {
	index := bfc.hash(key) % uint64(bfc.size)
	bfc.mu.RLock()
	defer bfc.mu.RUnlock()
	return bfc.array[index] == 0
}

// CheckCount checks the count for the given key.
func (bfc *BloomFilterCounter) CheckCount(key string) int {
	index := bfc.hash(key) % uint64(bfc.size)
	bfc.mu.RLock()
	defer bfc.mu.RUnlock()
	return int(bfc.array[index])
}

// hash uses a more robust hash function (FNV-1a)
func (bfc *BloomFilterCounter) hash(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}
